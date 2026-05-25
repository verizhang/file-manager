package server

import (
	"fmt"
	"net"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"

	grpcHandler "github.com/verizhang/file-manager/internal/handler"
	"github.com/verizhang/file-manager/internal/interceptor"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RunGRPCServer(
	port int,
	fileHandler *grpcHandler.FileHandler,
	logger *zap.Logger,
) (*grpc.Server, error) {
	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf(":%d", port),
	)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryRequestLogging(logger)),
	)
	filev1.RegisterFileServiceServer(
		grpcServer,
		fileHandler,
	)
	go func() {
		logger.Info(
			"grpc server started",
			zap.Int("port", port),
		)

		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal(
				"failed serve grpc server",
				zap.Error(err),
			)
		}
	}()

	return grpcServer, nil
}
