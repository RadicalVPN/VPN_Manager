package main

import (
	"radicalvpn/vpn-manager/logger"
	memstats "radicalvpn/vpn-manager/mem_stats"
	pingcheck "radicalvpn/vpn-manager/ping_check"
	privacyfirewall "radicalvpn/vpn-manager/privacy_firewall"
	publishqueue "radicalvpn/vpn-manager/publish_queue"
	wgparser "radicalvpn/vpn-manager/wg_parser"
	"runtime"
)

func main() {
	if runtime.GOOS != "linux" {
		logger.Warning.Println("This program should be run on Linux")
	}

	wgparser.Start()
	publishqueue.Start()
	pingcheck.Start()
	privacyfirewall.Start()
	memstats.Start()

	// keep the manager running
	select {}
}
