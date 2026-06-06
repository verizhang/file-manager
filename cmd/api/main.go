package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	configs "github.com/verizhang/file-manager/internal/config"
	"github.com/verizhang/file-manager/internal/database"
	grpcHandler "github.com/verizhang/file-manager/internal/handler"
	"github.com/verizhang/file-manager/internal/logger"
	"github.com/verizhang/file-manager/internal/messaging/rabbitmq"
	"github.com/verizhang/file-manager/internal/repository/mysql"
	"github.com/verizhang/file-manager/internal/server"
	"github.com/verizhang/file-manager/internal/service"
	"github.com/verizhang/file-manager/internal/storage/s3"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	//
	// =====================================================
	// LOAD ENV
	// =====================================================
	//
	err := godotenv.Load()
	if err != nil {
		log.Fatal("no .env file found")
	}

	//
	// =====================================================
	// CONFIG
	// =====================================================
	//
	cfg := configs.Load()

	//
	// =====================================================
	// LOGGER
	// =====================================================
	//

	logger.Init(cfg.App.Debug)

	defer logger.Log.Sync()

	logger.Log.Info("starting application")

	//
	// =====================================================
	// DATABASE
	// =====================================================
	//

	db, err := database.NewMySQLConnection(
		cfg.DB,
	)

	if err != nil {
		logger.Log.Fatal(
			"failed connect database",
			zap.Error(err),
		)
	}

	logger.Log.Info("database connected")

	//
	// =====================================================
	// MINIO / S3
	// =====================================================
	//
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

	//
	// =====================================================
	// REPOSITORY
	// =====================================================
	//

	fileRepository := mysql.NewFileRepository(
		db,
	)

	//
	// =====================================================
	// Messaging
	// =====================================================
	//

	messaging, err := rabbitmq.NewMessaging(cfg.RabbitMQ)
	if err != nil {
		logger.Log.Fatal("failed init rabbitmq client", zap.Error(err))
	}

	//
	// =====================================================
	// SERVICE
	// =====================================================
	//

	fileService := service.NewFileService(
		cfg,
		logger.Log,
		storageClient,
		fileRepository,
		messaging,
	)

	//
	// =====================================================
	// HANDLER
	// =====================================================
	//

	fileHandler := grpcHandler.NewFileHandler(
		logger.Log,
		fileService,
	)

	//
	// =====================================================
	// GRPC SERVER
	// =====================================================
	//

	grpcServer, err := server.RunGRPCServer(
		cfg.App.GRPCPort,
		fileHandler,
		logger.Log,
	)

	if err != nil {
		logger.Log.Fatal(
			"failed start grpc server",
			zap.Error(err),
		)
	}

	//
	// =====================================================
	// GATEWAY SERVER
	// =====================================================
	//

	httpServer, err := server.RunGatewayServer(
		cfg.App.HTTPPort,
		cfg.App.GRPCPort,
		logger.Log,
	)

	if err != nil {
		logger.Log.Fatal(
			"failed start gateway server",
			zap.Error(err),
		)
	}

	//
	// =====================================================
	// GRACEFUL SHUTDOWN
	// =====================================================
	//

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	logger.Log.Info("shutting down application")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Error(
			"failed shutdown http server",
			zap.Error(err),
		)
	}

	logger.Log.Info("application stopped")
}
