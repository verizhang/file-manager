package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	filev1 "github.com/verizhang/file-manager/gen/go/file/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

func RunGatewayServer(
	httpPort int,
	grpcPort int,
	logger *zap.Logger,
) (*http.Server, error) {
	ctx := context.Background()
	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			lowerKey := strings.ToLower(key)
			switch lowerKey {
			case "x-request-id", "request-id", "x-correlation-id":
				return lowerKey, true
			default:
				return runtime.DefaultHeaderMatcher(key)
			}
		}),
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		),
	)
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
