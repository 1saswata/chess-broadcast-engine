package server

import (
	"context"
	"strings"

	"github.com/1saswata/chess-broadcast-engine/internal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var a grpc.UnaryServerInterceptor

func AuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated,
			"metadata is not provided")
	}
	authHeader := md["authorization"]
	if len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated,
			"authorization token is not provided")
	}
	token := strings.TrimPrefix(authHeader[0], "Bearer")
	claims, err := auth.ValidateToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	if (*claims)["role"] != "grandmaster" {
		return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
	}
	return handler(ctx, req)
}
