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

	// TODO:
	// initialize minio client

	//
	// service
	//

	fileService := service.NewFileService(
		fileRepository,
		nil,
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