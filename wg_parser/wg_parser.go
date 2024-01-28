package wgparser

import (
	"fmt"
	"radicalvpn/vpn-manager/cli"
	"strconv"
	"strings"

	lo "github.com/samber/lo"
)

type WireguardInfo struct {
	publicKey string
	preSharedKey string
	endpoint string
	allowedIps []string
	latestHandshakeAt string
	transferRx int
	transferTx int
}

func GetParsedWireguardInfo() []WireguardInfo  {
	data := getInfoDump()
	trimmed := strings.TrimSpace(data)
	split := strings.Split(trimmed, "\n")
	slice := lo.Slice(split, 1, len(split) - 1)

	return lo.FilterMap(slice, func (line string, index int) (WireguardInfo, bool) {
		split := strings.Split(line, "\t")
		rx, err := strconv.Atoi(split[5])
		tx, err := strconv.Atoi(split[6])
		
		if err != nil {
			fmt.Println("[ERROR] Failed to parse rx/tx")
			return WireguardInfo{}, false
		}

		return WireguardInfo{
			publicKey: split[0],
			preSharedKey: split[1],
			endpoint: split[2],
			allowedIps: strings.Split(split[3], ","),
			latestHandshakeAt: split[4],
			transferRx: rx,
			transferTx: tx,
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

