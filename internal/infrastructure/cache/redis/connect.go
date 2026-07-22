package redisconnect

import (
	"context"
	"fmt"

	"github.com/chishkin-afk/learning_notes/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

// Connect opens connection with redis cache
//
// This function establishes a connection to the Redis cache
// and pings it, if something goes wrong, an error is returned.
func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(getOpts(cfg))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping client: %w", err)
	}

	return client, nil
}

func getOpts(cfg *config.Config) *redis.Options {
	return &redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Cache.Redis.Host,
			cfg.Cache.Redis.Port,
		),
		Username: cfg.Cache.Redis.Auth.Username,
		Password: cfg.Cache.Redis.Auth.Password,
	}
}
