package proxy

import (
	"testing"
)

func TestGetProxyLatency(t *testing.T) {
	latency, err := GetProxyLatency("vless://example.com")
	if latency == -1 {
		t.Errorf("latency test failed: %s", err)
	}
	latency, err = GetProxyLatency("vless://notareallink.fake")
	if latency != -1 {
		t.Errorf("latency test accepts fake link: latency %d", latency)
	}
}
