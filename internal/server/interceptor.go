package server

import (
	"context"
	"strings"

	"github.com/1saswata/chess-broadcast-engine/internal/auth"
	"github.com/1saswata/chess-broadcast-engine/internal/cache"
	"github.com/1saswata/chess-broadcast-engine/internal/pb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewAuthInterceptor(rc *cache.RedisCache) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
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
		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := auth.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		if claims["role"] != "grandmaster" {
			return nil, status.Errorf(codes.PermissionDenied, "unauthorized")
		}
		id, ok := claims["id"].(string)
		if !ok || id == "" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid user id")
		}
		allow, err := rc.AllowRequest(ctx, id, 2, 1)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "internal error")
		}
		if !allow {
			return nil, status.Errorf(codes.ResourceExhausted,
				"rate limit exceeded")
		}
		moveReq, ok := req.(*pb.Move)
		if ok {
			assignedColor, err := rc.GetPlayerColor(ctx, moveReq.MatchId, id)
			if err != nil {
				if err == redis.Nil {
					return nil, status.Errorf(codes.PermissionDenied,
						"not authorized")
				}
				return nil, status.Errorf(codes.Internal, "failed to  verify")
			}
			if assignedColor != moveReq.CurrentPlayer.String() {
				return nil, status.Errorf(codes.PermissionDenied, "wrong color")
			}
		}
		return handler(ctx, req)
	}
}
