package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
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

func (rc *RedisCache) AppendMove(ctx context.Context, matchID int32, move []byte) error {
	ctx, span := otel.Tracer("redis-cache").Start(ctx, "AppendMove")
	defer span.End()
	key := fmt.Sprintf("match:%d:latest", matchID)
	pipe := rc.client.Pipeline()
	pipe.RPush(ctx, key, move)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (rc *RedisCache) GetMoveHistory(ctx context.Context, matchID int32) ([][]byte, error) {
	key := fmt.Sprintf("match:%d:latest", matchID)
	s, err := rc.client.LRange(ctx, key, 0, -1).Result()
	b := make([][]byte, len(s))
	if err != nil {
		return nil, err
	}
	for i, msg := range s {
		b[i] = []byte(msg)
	}
	return b, nil
}

func (rc *RedisCache) IncrementSequence(ctx context.Context, matchID int32) (int32, error) {
	key := fmt.Sprintf("match:%d:sequence", matchID)
	seq, err := rc.client.Incr(ctx, key).Result()
	return int32(seq), err
}

func (rc *RedisCache) Close() error {
	return rc.client.Close()
}
