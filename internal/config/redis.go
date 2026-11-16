package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis() {
	addr := strings.TrimSpace(
		os.Getenv("REDIS_ADDR"),
	)

	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")

	dbNumber := 0

	if value := os.Getenv("REDIS_DB"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			dbNumber = parsed
		}
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNumber,
	})
}

func Ctx() context.Context {
	return context.Background()
}
