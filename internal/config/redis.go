// internal/config/redis.go
package config

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func Ctx() context.Context {
	return context.Background()
}
