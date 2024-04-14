package redis

import (
	"context"
	"fmt"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/realPointer/banners/pkg/redis"
)

type CacheRepo struct {
	*redis.Redis
}

func NewCacheRepo(rdb *redis.Redis) *CacheRepo {
	return &CacheRepo{
		Redis: rdb,
	}
}

func (r *CacheRepo) Get(ctx context.Context, key string) (string, error) {
	value, err := r.Client.Get(ctx, key).Result()
	if err == redisv8.Nil {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("redis - Get - r.Client.Get: %w", err)
	}
	return value, nil
}

func (r *CacheRepo) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := r.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis - Set - r.Client.Set: %w", err)
	}
	return nil
}
