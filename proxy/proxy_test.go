package proxy

import (
	"testing"
	"time"
)

func TestGetProxyLatency(t *testing.T) {
	latency, err := GetProxyLatency("vless://example.com", 60*time.Second)
	if latency == -1 {
		t.Errorf("latency test failed: %s", err)
	}
	latency, err = GetProxyLatency("vless://notareallink.fake", 60*time.Second)
	if latency != -1 {
		t.Errorf("latency test accepts fake link: latency %d", latency)
	}
}
