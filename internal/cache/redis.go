package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
		Password: "",
	})
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}
	return &RedisCache{client: rdb}, nil
}

func (rc *RedisCache) SetLatestMove(ctx context.Context, matchID int32, move []byte) error {
	key := fmt.Sprintf("match:%d:latest", matchID)
	return rc.client.Set(ctx, key, move, 24*time.Hour).Err()
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
