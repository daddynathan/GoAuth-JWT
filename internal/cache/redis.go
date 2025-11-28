package cache

import (
	"context"
	"fmt"
	"friend-help/internal/errs"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Close() error
}

type RedisService struct {
	client *redis.Client
}

func NewRedisService() (*RedisService, error) {
	addr := os.Getenv("REDIS_ADDR")
	pass := os.Getenv("REDIS_PASS")
	if addr == "" || pass == "" {
		return nil, fmt.Errorf("REDIS_ADDR or REDIS_PASS not set")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       1,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrFailedToPingRedis, err)
	}
	return &RedisService{client: rdb}, nil
}

func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(ctx, key, value, expiration)
}

func (r *RedisService) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

func (r *RedisService) Close() error {
	return r.client.Close()
}
