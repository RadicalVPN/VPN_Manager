package main

import (
	"fmt"
	publishqueue "radicalvpn/vpn-manager/publish_queue"
	wgparser "radicalvpn/vpn-manager/wg_parser"
	"runtime"
)

func main() {
	if runtime.GOOS != "linux" {
		fmt.Println("[WARN] This program should be run on Linux")
	}

	wgparser.StartTicker()
	publishqueue.Start()

	// keep the manager running
	select {}
}
