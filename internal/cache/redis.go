package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) (*RedisCache, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return &RedisCache{client: rdb}, nil
}

func (rc *RedisCache) SetLatestMove(ctx context.Context, matchID int32, move []byte) error {
	_ = fmt.Sprintf("match:%d:latest", matchID)
	return nil
}
