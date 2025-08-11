package cache

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPasswd := os.Getenv("REDIS_PASSWD")

	Client = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPasswd,
		DB:       0,
	})

	if err := Client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Error pinging redis: %v", err)
	}

	log.Printf("Connected to redis")
}

func CloseRedis() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			log.Printf("Error closing redis connection: %v", err)
		}
		log.Printf("Redis connection closed")
	}
}
