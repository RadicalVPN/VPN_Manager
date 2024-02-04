package privacyfirewall

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"radicalvpn/vpn-manager/logger"
	"radicalvpn/vpn-manager/redis"
	"regexp"
	"strings"
	"time"

	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

var interfaceRegex, _ = regexp.Compile("dns[0-9]*")
var httpPort = 4000

type PrivacyFirewallResponse struct {
	TotalRequests        map[string]int `json:"total"`
	TotalBlockedRequests map[string]int `json:"blocked"`
}

func StartTicker() {
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				ComputeMetrics()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func ComputeMetrics() {
	data := computePrivacyFirewallStats()
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error.Println("failed to get hostname", err)
		return
	}

	setErr := redis.GetClient().JSONSet(context.Background(), fmt.Sprintf("privacy_firewall_stats:%s", hostname), "$", data).Err()
	if setErr != nil {
		logger.Error.Println("failed to set redis data", setErr)
		return
	}

	logger.Info.Println("Updated privacy firewall stats")
}

func computePrivacyFirewallStats() PrivacyFirewallResponse {
	ips := getAllPrivacyFirewallIps()
	parser := expfmt.TextParser{}
	var totalRequest = make(map[string]int)
	var totalBlockedRequests = make(map[string]int)

	for _, ip := range ips {
		data, err := getMetricsFromIp(ip)
		if err != nil {
			logger.Error.Println("failed to get metrics from", ip, err)
			continue
		}

		mf, err := parser.TextToMetricFamilies(strings.NewReader(data))
		if err != nil {
			logger.Error.Println("failed to parse prometheus metrics from", ip, err)
		}

		for k, v := range mf {
			if k == "blocky_query_total" {
				for client, count := range computeRequestCountFromMetric(v) {
					if _, ok := totalRequest[client]; ok {
						totalRequest[client] = totalRequest[client] + count
					} else {
						totalRequest[client] = count
					}
				}
			}

			if k == "blocky_query_blocked_total" {
				for client, count := range computeRequestCountFromMetric(v) {
					if _, ok := totalBlockedRequests[client]; ok {
						totalBlockedRequests[client] = totalBlockedRequests[client] + count
					} else {
						totalBlockedRequests[client] = count
					}
				}
			}
		}
	}

	return PrivacyFirewallResponse{
		TotalRequests:        totalRequest,
		TotalBlockedRequests: totalBlockedRequests,
	}
}

func getAllPrivacyFirewallIps() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	ips := make([]string, 0)
	for _, iface := range ifaces {
		if interfaceRegex.MatchString(iface.Name) {
			addrs, err := iface.Addrs()
			if err != nil {
				panic(err)
			}

			ips = append(ips, addrs[0].(*net.IPNet).IP.String())
		}
	}

	return ips
}

func getMetricsFromIp(ip string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/metrics", ip, httpPort))
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func computeRequestCountFromMetric(metric *io_prometheus_client.MetricFamily) map[string]int {
	result := make(map[string]int)

	for _, m := range metric.Metric {
		client := m.GetLabel()[0].GetValue()
		count := m.GetCounter().GetValue()

		if _, ok := result[client]; ok {
			result[client] = result[client] + int(count)
		} else {
			result[client] = int(count)
		}
	}

	return result
}
