package main

import (
	"fmt"
	wgparser "radicalvpn/vpn-manager/wg_parser"
	"runtime"
)

func main() {
	if runtime.GOOS != "linux" {
		fmt.Println("[WARN] This program should be run on Linux")
	}

	for true {
		wgparser.WireDataToRedis()
	}
}
