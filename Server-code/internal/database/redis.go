package database

import (
	"context"
	"fmt"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

func InitRedis(cfg *config.Config) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:            cfg.RedisAddr(),
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		PoolSize:        1,
		MinIdleConns:    0,
		DialTimeout:     time.Duration(cfg.Redis.DialTimeoutSeconds) * time.Second,
		ReadTimeout:     time.Duration(cfg.Redis.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:    time.Duration(cfg.Redis.WriteTimeoutSeconds) * time.Second,
		MaxRetries:      0,
		ConnMaxLifetime: 0,
		PoolTimeout:     1 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected successfully",
		zap.String("addr", cfg.RedisAddr()),
		zap.Int("db", cfg.Redis.DB),
	)

	return nil
}

func CloseRedis() {
	if RDB != nil {
		_ = RDB.Close()
		logger.Info("Redis connection closed")
	}
}
