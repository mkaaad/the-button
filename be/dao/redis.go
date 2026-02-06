package dao

import (
	"context"
	"log"

	"button/config"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

func InitRedis() {
	// 初始化 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("无法连接至 Redis: %v", err)
	}
	Rdb = rdb
}
