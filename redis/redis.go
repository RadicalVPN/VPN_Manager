package redis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

// cache the client to avoid multiple connections
var client *redis.Client

func GetClient() redis.Client {
	if client == nil {
		client = GetNewClient()
	}

	return *client
}

func GetNewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     getRedisUrl(),
		Password: "",
		DB:       0,
	})
}

func getRedisUrl() string {
	url := os.Getenv("REDIS_URL")

	if url == "" {
		panic("[ERROR] REDIS_URL environment variable not set")
	}


	return url
}
