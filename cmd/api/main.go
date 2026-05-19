package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/verizhang/file-manager/internal/config"

	"github.com/verizhang/file-manager/internal/database"

	grpcHandler "github.com/verizhang/file-manager/internal/handler"

	"github.com/verizhang/file-manager/internal/logger"

	"github.com/verizhang/file-manager/internal/repository"

	"github.com/verizhang/file-manager/internal/server"

	"github.com/verizhang/file-manager/internal/service"

	minioStorage "github.com/verizhang/file-manager/internal/storage"

	"go.uber.org/zap"

	"github.com/joho/godotenv"

	"log"
)

func main() {
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
	// MINIO
	// =====================================================
	//

	minioClient, err := minio.New(
		cfg.S3.Endpoint,
		&minio.Options{
			Creds: credentials.NewStaticV4(
				cfg.S3.AccessKey,
				cfg.S3.SecretKey,
				"",
			),
			Secure: cfg.S3.UseSSL,
			Region: cfg.S3.Region,
		},
	)

	if err != nil {
		logger.Log.Fatal(
			"failed initialize minio client",
			zap.Error(err),
		)
	}

	logger.Log.Info("minio initialized")

	minioProvider := minioStorage.New(
		minioClient,
		cfg.S3.Bucket,
	)

	//
	// =====================================================
	// REPOSITORY
	// =====================================================
	//

	fileRepository := repository.NewFileRepository(
		db,
	)

	//
	// =====================================================
	// SERVICE
	// =====================================================
	//

	fileService := service.NewFileService(
		fileRepository,
		minioProvider,
	)

	//
	// =====================================================
	// HANDLER
	// =====================================================
	//

	fileHandler := grpcHandler.NewFileHandler(
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