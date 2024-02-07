package serverloadcalculatorgo

import (
	"context"
	"fmt"
	"os"
	"radicalvpn/vpn-manager/helpers/redis"
	"strconv"
	"strings"
	"time"
)

const (
	rxFilePath = "/sys/class/net/wg0/statistics/rx_bytes"
	txFilePath = "/sys/class/net/wg0/statistics/tx_bytes"
	maxLoad    = 2000.0 // 2 Gigabit
)

type InterfaceLoad struct {
	Rx uint64
	Tx uint64
}

type LoadHistory struct {
	RxMbps float64
	TxMbps float64
}

var lastLoad InterfaceLoad
var loadHistory = make([]LoadHistory, 0)

func Start() {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				computeLoad()
			}
		}
	}()
}

func computeLoad() {
	rxBytes, txBytes := getRawStats()

	//skip on first run
	if lastLoad.Rx == 0 {
		updateLastLoad(rxBytes, txBytes)
		return
	}

	// get current load
	currentLoad := InterfaceLoad{
		Rx: rxBytes - lastLoad.Rx,
		Tx: txBytes - lastLoad.Tx,
	}

	// convert to Mbit/s
	rxMbps := float64(currentLoad.Rx) * 8 / 1000000
	txMbps := float64(currentLoad.Tx) * 8 / 1000000
	updateHistory(rxMbps, txMbps)

	rxAvg, txAvg := getAverageLoad()

	// calculate load percentage
	loadPercent := fmt.Sprintf("%.2f", ((rxAvg+txAvg)/maxLoad)*100)

	redis.GetClient().Set(context.Background(), getRedisKey(), loadPercent, 30*time.Second)

	updateLastLoad(rxBytes, txBytes)
}

func getAverageLoad() (float64, float64) {
	var rxSum float64
	var txSum float64

	for _, load := range loadHistory {
		rxSum += load.RxMbps
		txSum += load.TxMbps
	}

	rxAvg := rxSum / float64(len(loadHistory))
	txAvg := txSum / float64(len(loadHistory))

	return rxAvg, txAvg
}

func updateHistory(rxMbps float64, txMbps float64) {
	loadHistory = append(loadHistory, LoadHistory{
		RxMbps: rxMbps,
		TxMbps: txMbps,
	})

	// only keep the last 60 seconds
	if len(loadHistory) > 60 {
		loadHistory = loadHistory[1:]
	}
}

func updateLastLoad(rx uint64, tx uint64) {
	lastLoad = InterfaceLoad{
		Rx: rx,
		Tx: tx,
	}
}

func getRawStats() (uint64, uint64) {
	rxBytes, err := os.ReadFile(rxFilePath)
	if err != nil {
		fmt.Println("failed to read rx bytes", err)
		return 0, 0
	}

	txBytes, err := os.ReadFile(txFilePath)
	if err != nil {
		fmt.Println("failed to read tx bytes", err)
		return 0, 0
	}

	rxBytesInt, err := strconv.ParseUint(strings.Replace(string(rxBytes), "\n", "", -1), 10, 64)
	if err != nil {
		fmt.Println("failed to parse rx bytes", err)
		return 0, 0
	}

	txBytesInt, err := strconv.ParseUint(strings.Replace(string(txBytes), "\n", "", -1), 10, 64)
	if err != nil {
		fmt.Println("failed to parse tx bytes", err)
		return 0, 0
	}

	return rxBytesInt, txBytesInt
}

func getRedisKey() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("server-load-percent:%s", hostname)
}
