package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func RunGatewayServer(
	httpPort int,
	grpcPort int,
	logger *zap.Logger,
) (*http.Server, error) {

	ctx := context.Background()

	mux := runtime.NewServeMux()

	err := filev1.RegisterFileServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", grpcPort),
		[]grpc.DialOption{
			grpc.WithTransportCredentials(
				insecure.NewCredentials(),
			),
		},
	)

	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", httpPort),
		Handler: mux,
	}

	go func() {
		logger.Info(
			"http gateway started",
			zap.Int("port", httpPort),
		)

		if err := httpServer.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.Fatal(
				"failed start http gateway",
				zap.Error(err),
			)
		}
	}()

	return httpServer, nil
}