package redis

import (
	"context"
	"os"
	"radicalvpn/vpn-manager/logger"

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
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			logger.Debug.Println("connected to redis", cn.ClientID(ctx))
			return nil
		},
	})
}

func getRedisUrl() string {
	url := os.Getenv("REDIS_URL")

	if url == "" {
		logger.Error.Fatalln("[ERROR] REDIS_URL environment variable not set")
	}

	return url
}
