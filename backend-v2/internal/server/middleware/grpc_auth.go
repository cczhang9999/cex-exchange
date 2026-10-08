package middleware

import (
	"context"
	"strings"

	"backend-v2/internal/pkg/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserIDKey is the context key under which the authenticated user ID is stored.
type contextKey string

const UserIDKey contextKey = "uid"

// GrpcAuthInterceptor creates a gRPC server interceptor that authenticates
// the caller via a JWT in the gRPC metadata "authorization" header
// (format: "Bearer <token>"), validates it, and stores the user ID in the
// request context.
//
// The stored user ID can be retrieved downstream via FromContext(ctx).
func GrpcAuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		uid, err := authenticateFromMetadata(ctx, secret)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "auth failed: %v", err)
		}
		ctx = context.WithValue(ctx, UserIDKey, uid)
		return handler(ctx, req)
	}
}

// authenticateFromMetadata extracts and validates the JWT from gRPC metadata.
func authenticateFromMetadata(ctx context.Context, secret string) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	auths := md.Get("authorization")
	if len(auths) == 0 {
		// Fallback: check "token" key in metadata
		auths = md.Get("token")
	}
	if len(auths) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	authHeader := auths[0]
	parts := strings.SplitN(authHeader, " ", 2)
	tokenStr := authHeader
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		tokenStr = parts[1]
	}

	claims, err := jwt.ParseToken(tokenStr, secret)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims.UserID, nil
}

// FromContext extracts the authenticated user ID from the context.
// This is populated by the GrpcAuthInterceptor for incoming gRPC requests.
func FromContext(ctx context.Context) (uint64, bool) {
	uid, ok := ctx.Value(UserIDKey).(uint64)
	return uid, ok
}
