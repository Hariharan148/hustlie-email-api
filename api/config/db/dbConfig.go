package db


import (
	"github.com/go-redis/redis/v8"
	"context"
	"os"
)

var Ctx = context.Background()


func RedisClient(dbInt int) *redis.Client{

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		Password:os.Getenv("REDIS_PASS"),
		DB: dbInt,
	})

	return rdb
}