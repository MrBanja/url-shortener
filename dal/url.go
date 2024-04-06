package dal

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var NotFoundError = fmt.Errorf("not found")

type DAL struct {
	client *redis.Client
}

func MustNew(ctx context.Context, connStr string) *DAL {
	opt, err := redis.ParseURL(connStr)
	if err != nil {
		panic(fmt.Errorf("parse redis url: %w", err))
	}
	client := redis.NewClient(opt)
	if err = client.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("ping redis: %w", err))
	}
	return &DAL{
		client: client,
	}
}

func (d *DAL) Close() error {
	return d.client.Close()
}

func (d *DAL) Ping(ctx context.Context) (string, error) {
	return d.client.Ping(ctx).Result()
}

func (d *DAL) Get(ctx context.Context, short string) (string, error) {
	url, err := d.client.Get(ctx, short).Result()
	if errors.Is(err, redis.Nil) {
		return "", NotFoundError
	}
	return url, err
}

func (d *DAL) Set(ctx context.Context, short, long string) error {
	return d.client.Set(ctx, short, long, 0).Err()
}
