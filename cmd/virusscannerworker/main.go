package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	configs "github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/consumer"
	"github.com/verizhang/file-manager/internal/repository/mysql"
	"github.com/verizhang/file-manager/internal/service"
	database "github.com/verizhang/file-manager/pkg/database/mysql"
	"github.com/verizhang/file-manager/pkg/logger"
	"github.com/verizhang/file-manager/pkg/messaging/rabbitmq"
	"github.com/verizhang/file-manager/pkg/storage/s3"
	"github.com/verizhang/file-manager/pkg/virusscanner/clamav"
	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatal("no .env file found")
	}

	cfg := configs.Load()
	if !cfg.ClamAV.Enabled {
		log.Fatal("clamav scanner is disabled")
	}

	logger.Init(cfg.App.Debug)
	defer logger.Log.Sync()

	logger.Log.Info("starting virus scanner worker")

	db, err := database.NewMySQLConnection(cfg.DB)
	if err != nil {
		logger.Log.Fatal("failed connect database", zap.Error(err))
	}

	s3Client, err := s3.NewClient(
		ctx,
		s3.Config{
			Endpoint:  cfg.S3.Endpoint,
			Region:    cfg.S3.Region,
			AccessKey: cfg.S3.AccessKey,
			SecretKey: cfg.S3.SecretKey,
			UseSSL:    cfg.S3.UseSSL,
		},
	)
	if err != nil {
		logger.Log.Fatal("failed init s3 client", zap.Error(err))
	}

	storageClient := s3.NewStorage(s3Client)
	fileRepository := mysql.NewFileRepository(db)

	messagingClient, err := rabbitmq.NewMessaging(cfg.RabbitMQ)
	if err != nil {
		logger.Log.Fatal("failed init rabbitmq client", zap.Error(err))
	}
	defer messagingClient.Close()

	virusScanner := clamav.NewScanner(cfg.ClamAV)

	fileService := service.NewFileService(
		cfg,
		logger.Log,
		storageClient,
		fileRepository,
		messagingClient,
		virusScanner,
	)

	handler := consumer.NewHandler(logger.Log, fileService)

	if err := messagingClient.Subscribe(ctx, service.COMPLETE_UPLOAD_TOPIC, handler.VirusScanner); err != nil {
		logger.Log.Fatal("failed subscribe virus scanner worker", zap.Error(err))
	}

	logger.Log.Info(
		"virus scanner worker subscribed",
		zap.String("topic", service.COMPLETE_UPLOAD_TOPIC),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Log.Info("shutting down virus scanner worker")
	cancel()

	logger.Log.Info("virus scanner worker stopped")
}
