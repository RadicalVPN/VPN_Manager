package redis

import (
	"github.com/redis/go-redis/v9"
)

// cache the client to avoid multiple connections
var client *redis.Client

func GetClient() redis.Client {
	if client == nil {
		client = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		})
	}

	return *client
}
