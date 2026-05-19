package main

import (
	"github.com/verizhang/file-manager/internal/config"
	grpcHandler "github.com/verizhang/file-manager/internal/handler"
	"github.com/verizhang/file-manager/internal/logger"
	"github.com/verizhang/file-manager/internal/repository"
	"github.com/verizhang/file-manager/internal/service"
)

func main() {
	cfg := configs.Load()

	logger.Init(cfg.App.Debug)

	//
	// repository
	//

	fileRepository := repository.NewFileRepository()

	//
	// storage
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

	logger.Log.Info("minio client initialized")

	storage := storage.New(minioClient, cfg.S3.Bucket)

	//
	// service
	//

	fileService := service.NewFileService(
		fileRepository,
		storage,
	)

	//
	// handler
	//

	_ = grpcHandler.NewFileHandler(
		fileService,
	)

	//
	// TODO:
	// start grpc server
	// start grpc gateway
	//
}