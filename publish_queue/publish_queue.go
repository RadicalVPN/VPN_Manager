package publishqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"radicalvpn/vpn-manager/cli"
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

	fmt.Println("[INFO] Started publish queue", queueKey)

	for {
		redisData := redis.GetNewClient().BRPop(context.Background(), 0, queueKey)

		if redisData.Err() != nil {
			fmt.Println("[ERROR] Error reading from redis", redisData.Err())
			return
		}

		data := redisData.Val()[1]
		message := parseQueueMessage(data)

		fmt.Println("[INFO] handling queue message from redis")

		f, err := os.Create(wireguardConfigPath)
		if err != nil {
			fmt.Println("[ERROR] failed to open wg0.conf", err)
		}

		count, err := f.WriteString(message.Config)
		if err != nil {
			fmt.Println("[ERROR] failed to write to wg0.conf", err)
		}

		f.Close()

		fmt.Println("[INFO] wrote", count, "bytes to wg0.conf")

		if _, err := cli.Exec("wg", "syncconf", "wg0", wireguardConfigPath); err != nil {
			fmt.Println("[ERROR] failed to update wg0.conf", err)
		}
	}
}

func parseQueueMessage(data string) PublishQueueMessage {
	message := PublishQueueMessage{}

	err := json.Unmarshal([]byte(data), &message)

	if err != nil {
		fmt.Println("[ERROR] Error parsing queue message", err)
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
