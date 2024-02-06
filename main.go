package main

import (
	"radicalvpn/vpn-manager/helpers/logger"
	memstats "radicalvpn/vpn-manager/modules/mem_stats"
	pingcheck "radicalvpn/vpn-manager/modules/ping_check"
	privacyfirewall "radicalvpn/vpn-manager/modules/privacy_firewall"
	publishqueue "radicalvpn/vpn-manager/modules/publish_queue"
	wgparser "radicalvpn/vpn-manager/modules/wg_parser"
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
