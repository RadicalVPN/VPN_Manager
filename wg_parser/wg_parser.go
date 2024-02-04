package wgparser

import (
	"context"
	"fmt"
	"os"
	"radicalvpn/vpn-manager/cli"
	"radicalvpn/vpn-manager/redis"
	"strconv"
	"strings"
	"time"

	lo "github.com/samber/lo"
)

type TrafficInfo struct {
	rx int
	tx int
}

type WireguardInfo struct {
	publicKey         string
	preSharedKey      string
	endpoint          string
	allowedIps        []string
	latestHandshakeAt string
	transferRx        int
	transferTx        int
}

type RedisResult struct {
	PublicKey string `json:"publicKey"`
	Tx        int    `json:"transferTx"`
	Rx        int    `json:"transferRx"`
	Connected bool   `json:"connected"`
}

var lastStats = make(map[string]TrafficInfo)

func StartTicker() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				ParseAndWriteData()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func ParseAndWriteData() {
	parsed := GetParsedWireguardInfo()
	hostname, err := os.Hostname()

	if len(parsed) == 0 {
		fmt.Println("[ERROR] No wireguard data to compute")
		return
	}

	if err != nil {
		fmt.Println("[ERROR] Failed to get hostname")
		return
	}

	currentStats := make(map[string]TrafficInfo)
	lo.ForEach(parsed, func(info WireguardInfo, index int) {
		currentStats[info.publicKey] = TrafficInfo{
			rx: info.transferRx,
			tx: info.transferTx,
		}
	})

	connectionStateKeys := lo.Map(parsed, func(info WireguardInfo, index int) string {
		return fmt.Sprintf("vpn_connection_state:%s", info.publicKey)
	})
	connectionStateResult, err := redis.GetClient().MGet(context.Background(), connectionStateKeys...).Result()
	if err != nil {
		fmt.Println("[ERROR] Failed to get connection states from redis", err)
		return
	}

	redisResults := make(map[string]RedisResult)
	lo.ForEach(parsed, func(vpn WireguardInfo, index int) {
		if _, ok := lastStats[vpn.publicKey]; !ok {
			return
		}

		tx := vpn.transferTx - lastStats[vpn.publicKey].tx
		rx := vpn.transferRx - lastStats[vpn.publicKey].rx
		connectionState := connectionStateResult[index]
		var connected bool

		if connectionState != nil {
			connected = true
		} else {
			lastTx := lastStats[vpn.publicKey].tx
			lastRx := lastStats[vpn.publicKey].rx

			connected = vpn.transferRx != lastRx || vpn.transferTx != lastTx

			if connected {
				setErr := redis.GetClient().Set(context.Background(), fmt.Sprintf("vpn_connection_state:%s", vpn.publicKey), "dummy", 30).Err()
				if setErr != nil {
					fmt.Println("[ERROR] Failed to set connection state", err)
					return
				}
			}
		}

		redisResults[vpn.publicKey] = RedisResult{
			PublicKey: vpn.publicKey,
			Tx:        tx,
			Rx:        rx,
			Connected: connected,
		}
	})

	fmt.Println(fmt.Sprintf("[INFO] Parsed %d wireguard connections", len(parsed)))

	setErr := redis.GetClient().JSONSet(context.Background(), fmt.Sprintf("vpn_stats:%s", hostname), "$", redisResults).Err()
	if setErr != nil {
		fmt.Println("[ERROR] Failed to set redis data", err)
		return
	}

	lastStats = currentStats
}

func GetParsedWireguardInfo() []WireguardInfo {
	data := getInfoDump()
	trimmed := strings.TrimSpace(data)
	split := strings.Split(trimmed, "\n")
	slice := lo.Slice(split, 1, len(split))

	return lo.FilterMap(slice, func(line string, index int) (WireguardInfo, bool) {
		split := strings.Split(line, "\t")
		rx, err := strconv.Atoi(split[5])
		tx, err := strconv.Atoi(split[6])

		if err != nil {
			fmt.Println("[ERROR] Failed to parse rx/tx")
			return WireguardInfo{}, false
		}

		return WireguardInfo{
			publicKey:         split[0],
			preSharedKey:      split[1],
			endpoint:          split[2],
			allowedIps:        strings.Split(split[3], ","),
			latestHandshakeAt: split[4],
			transferRx:        rx,
			transferTx:        tx,
		}, true
	})
}

func getInfoDump() string {
	result, err := cli.Exec("wg", "show", "wg0", "dump")
	if err != nil {
		fmt.Println("[ERROR] Failed to execute wg command")
		return ""
	}

	return result
}
