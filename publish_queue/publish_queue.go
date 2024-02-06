package publishqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"radicalvpn/vpn-manager/cli"
	"radicalvpn/vpn-manager/logger"
	"radicalvpn/vpn-manager/redis"
)

var wireguardConfigPath = "/etc/wireguard/wg0.conf"

type PublishQueueMessage struct {
	Config string `json:"config"`
}

func Start() {
	go start()
}

func start() {
	queueKey := getQueueKey()

	logger.Info.Println("Started publish queue:", queueKey)

	for {
		redisData := redis.GetNewClient().BRPop(context.Background(), 0, queueKey)

		if redisData.Err() != nil {
			logger.Error.Println("Error reading from redis", redisData.Err())
			return
		}

		data := redisData.Val()[1]
		message := parseQueueMessage(data)

		logger.Info.Println("handling queue message from redis")

		f, err := os.Create(wireguardConfigPath)
		if err != nil {
			logger.Error.Println("failed to open wg0.conf", err)
		}
		defer f.Close()

		count, err := f.WriteString(message.Config)
		if err != nil {
			logger.Error.Println("failed to write to wg0.conf", err)
		}



		logger.Info.Println("wrote", count, "bytes to wg0.conf")

		if _, err := cli.Exec("wg", "syncconf", "wg0", wireguardConfigPath); err != nil {
			logger.Error.Println("failed to update wg0.conf", err)
		}
	}
}

func parseQueueMessage(data string) PublishQueueMessage {
	message := PublishQueueMessage{}

	err := json.Unmarshal([]byte(data), &message)

	if err != nil {
		logger.Error.Println("Error parsing queue message", err)
	}

	return message
}

func getQueueKey() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("vpn_manager:publish_queue:%s", hostname)
}
