package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client *redis.Client
}

func New(url string, opts ...Option) (*Redis, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis - New - redis.ParseURL: %w", err)
	}

	rdb := &Redis{
		Client: redis.NewClient(opt),
	}

	for _, opt := range opts {
		opt(rdb)
	}

	if err := rdb.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis - New - rdb.Ping: %w", err)
	}

	return rdb, nil
}

func (r *Redis) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *Redis) Close() error {
	return r.Client.Close()
}
