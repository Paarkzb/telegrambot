package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"telegrambot/internal/config"
)

type RepositoryRedis struct {
	db *redis.Client
}

func New(config *config.Config) *RepositoryRedis {
	db := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisAddr, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	db.Ping(context.Background())

	return &RepositoryRedis{db: db}
}

func (r *RepositoryRedis) GetState(ctx context.Context, key string) (string, error) {
	return r.db.Get(ctx, key).Result()
}

func (r *RepositoryRedis) SetState(ctx context.Context, key string, value string) error {
	return r.db.Set(ctx, key, value, 0).Err()
}

func (r *RepositoryRedis) DeleteState(ctx context.Context, key string) error {
	return r.db.Del(ctx, key).Err()
}
