package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	// UserIDContextKey is the context key for the user ID
	UserIDContextKey contextKey = "x-user-id"
)

// AuthInterceptor returns a new unary server interceptor that authenticates requests.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var userID string

		// Check for gRPC metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if ids := md.Get("x-user-id"); len(ids) > 0 {
				userID = ids[0]
			}
		}

		if userID == "" {
			return nil, status.Errorf(codes.Unauthenticated, "X-User-ID header is required")
		}

		// Inject userID into context
		newCtx := context.WithValue(ctx, UserIDContextKey, userID)

		return handler(newCtx, req)
	}
}
