package interceptor

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerRequestID     = "x-request-id"
	headerAltRequestID  = "request-id"
	headerCorrelationID = "x-correlation-id"
)

func UnaryRequestLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestID := extractRequestID(ctx)
		_ = grpc.SetHeader(ctx, metadata.Pairs(headerRequestID, requestID))

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Milliseconds()

		code := status.Code(err)
		if err == nil {
			code = codes.OK
		}

		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", info.FullMethod),
			zap.String("grpc_code", code.String()),
			zap.Int64("duration", duration),
		}

		switch code {
		case codes.Internal,
			codes.Unavailable,
			codes.DeadlineExceeded,
			codes.DataLoss,
			codes.Unknown:
			logger.Error("grpc_request_completed", fields...)
		case codes.InvalidArgument,
			codes.NotFound,
			codes.AlreadyExists,
			codes.FailedPrecondition,
			codes.OutOfRange,
			codes.Aborted,
			codes.PermissionDenied,
			codes.Unauthenticated,
			codes.ResourceExhausted,
			codes.Canceled:
			logger.Warn("grpc_request_completed", fields...)
		default:
			logger.Info("grpc_request_completed", fields...)
		}

		return resp, err
	}
}

func extractRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if requestID := first(md.Get(headerRequestID)); requestID != "" {
			return requestID
		}
		if requestID := first(md.Get(headerAltRequestID)); requestID != "" {
			return requestID
		}
		if requestID := first(md.Get(headerCorrelationID)); requestID != "" {
			return requestID
		}
	}

	return uuid.NewString()
}

func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
