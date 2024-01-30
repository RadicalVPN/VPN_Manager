package pingcheck

import (
	"context"
	"fmt"
	"os"
	"radicalvpn/vpn-manager/redis"
)

func Start() {
	go start()
}

func start() {
	redis := redis.GetClient()
	channel := getPingChannelKey()
	ctx := context.Background()

	fmt.Println("[INFO] Started ping/pong", channel)

	subscriber := redis.Subscribe(ctx, channel)
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			fmt.Println("[ERROR] Error reading from redis", err)
			return
		}

		fmt.Println("[INFO] got ping message from control server", msg.Payload)
		redis.Publish(ctx, getPongChannelKey(), "")
	}
}

func getPingChannelKey() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("ping:%s", hostname)
}

func getPongChannelKey() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("pong:%s", hostname)
}