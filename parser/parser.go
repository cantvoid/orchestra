package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func ProxyToSingbox(proxyURL string) (map[string]interface{}, error) {
	var outbound map[string]interface{}
	var err error

	switch {
	case strings.HasPrefix(proxyURL, "vless://"):
		outbound, err = vlessToSingbox(proxyURL)
	case strings.HasPrefix(proxyURL, "vmess://"):
		outbound, err = vmessToSingbox(proxyURL)
	case strings.HasPrefix(proxyURL, "trojan://"):
		outbound, err = trojanToSingbox(proxyURL)
	}
	if err != nil {
		return nil, err
	}

	serverHostname := outbound["server"].(string)
	outbound["tag"] = "proxy"

	return map[string]interface{}{
		"log": map[string]interface{}{"level": "info", "timestamp": true},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "local_local", "type": "udp", "server": "223.5.5.5"},
				{"tag": "remote_dns", "type": "https", "server": "cloudflare-dns.com", "domain_resolver": "hosts_dns", "path": "/dns-query", "detour": "proxy"},
				{"tag": "direct_dns", "type": "https", "server": "dns.alidns.com", "domain_resolver": "hosts_dns", "path": "/dns-query"},
				{
					"tag":  "hosts_dns",
					"type": "hosts",
					"predefined": map[string][]string{
						"dns.google":                  {"8.8.8.8", "8.8.4.4"},
						"dns.alidns.com":              {"223.5.5.5", "223.6.6.6"},
						"one.one.one.one":             {"1.1.1.1", "1.0.0.1"},
						"cloudflare-dns.com":          {"104.16.249.249", "104.16.248.249"},
						"dns.cloudflare.com":          {"104.16.132.229", "104.16.133.229"},
						"dot.pub":                     {"1.12.12.12", "120.53.53.53"},
						"doh.pub":                     {"1.12.12.12", "120.53.53.53"},
						"dns.quad9.net":               {"9.9.9.9", "149.112.112.112"},
						"dns.yandex.net":              {"77.88.8.8", "77.88.8.1"},
						"dns.sb":                      {"185.222.222.222"},
						"dns.umbrella.com":            {"208.67.220.220", "208.67.222.222"},
						"dns.sse.cisco.com":           {"208.67.220.220", "208.67.222.222"},
						"engage.cloudflareclient.com": {"162.159.192.1"},
					},
				},
			},
			"rules": []map[string]interface{}{
				{"domain": []string{serverHostname}, "server": "local_local"},
				{"server": "hosts_dns", "ip_accept_any": true},
				{"server": "remote_dns", "clash_mode": "Global"},
				{"server": "direct_dns", "clash_mode": "Direct"},
				{"action": "predefined", "rcode": "NOTIMP", "query_type": []int{64, 65}},
			},
			"final":             "remote_dns",
			"independent_cache": true,
		},
		"inbounds": []map[string]interface{}{
			{
				"type":           "tun",
				"tag":            "tun-in",
				"interface_name": "singbox_tun",
				"address":        []string{"172.18.0.1/30"},
				"mtu":            9000,
				"auto_route":     true,
				"strict_route":   true,
				"stack":          "gvisor",
			},
		},
		"outbounds": []map[string]interface{}{
			outbound,
			{"type": "direct", "tag": "direct"},
		},
		"route": map[string]interface{}{
			"default_domain_resolver": map[string]interface{}{"server": "direct_dns", "strategy": ""},
			"auto_detect_interface":   true,
			"rules": []map[string]interface{}{
				{"port": []int{53}, "process_name": []string{"v2ray", "xray", "clash", "mihomo", "hysteria", "naive", "naiveproxy", "tuic-client", "tuic", "sing-box-client", "sing-box", "brook", "overtls"}, "action": "hijack-dns"},
				{"outbound": "direct", "process_name": []string{"v2ray", "xray", "clash", "mihomo", "hysteria", "naive", "naiveproxy", "tuic-client", "tuic", "sing-box-client", "sing-box", "brook", "overtls"}},
				{"action": "sniff"},
				{"protocol": []string{"dns"}, "action": "hijack-dns"},
				{"outbound": "direct", "clash_mode": "Direct"},
				{"outbound": "proxy", "clash_mode": "Global"},
				{"network": []string{"udp"}, "port": []int{443}, "action": "reject"},
				{"outbound": "direct", "ip_is_private": true},
				{"outbound": "proxy", "port_range": []string{"0:65535"}},
			},
			"final": "proxy",
		},
		"experimental": map[string]interface{}{
			"cache_file": map[string]interface{}{"enabled": true, "store_fakeip": false},
			"clash_api":  map[string]interface{}{"external_controller": "127.0.0.1:10814"},
		},
	}, nil
}

func vlessToSingbox(proxyURL string) (map[string]interface{}, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	outbound := map[string]interface{}{
		"type":            "vless",
		"server":          u.Hostname(),
		"server_port":     getPort(u),
		"uuid":            u.User.Username(),
		"flow":            q.Get("flow"),
		"packet_encoding": "xudp",
	}
	if q.Get("security") == "reality" {
		outbound["tls"] = map[string]interface{}{
			"enabled":     true,
			"server_name": getSNI(u),
			"insecure":    false,
			"utls": map[string]interface{}{
				"enabled":     true,
				"fingerprint": q.Get("fp"),
			},
			"reality": map[string]interface{}{
				"enabled":    true,
				"public_key": q.Get("pbk"),
				"short_id":   q.Get("sid"),
			},
		}
	}
	return outbound, nil
}

func vmessToSingbox(proxyURL string) (map[string]interface{}, error) {
	b64Part := strings.TrimPrefix(proxyURL, "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(fixPadding(b64Part))
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	json.Unmarshal(decoded, &data)
	return map[string]interface{}{
		"type":        "vmess",
		"server":      fmt.Sprint(data["add"]),
		"server_port": int(data["port"].(float64)),
		"uuid":        fmt.Sprint(data["id"]),
		"security":    "auto",
	}, nil
}

func trojanToSingbox(proxyURL string) (map[string]interface{}, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"type":        "trojan",
		"server":      u.Hostname(),
		"server_port": getPort(u),
		"password":    u.User.Username(),
	}, nil
}

func getPort(u *url.URL) int {
	p, _ := strconv.Atoi(u.Port())
	if p == 0 {
		return 443
	}
	return p
}

func getSNI(u *url.URL) string {
	if sni := u.Query().Get("sni"); sni != "" {
		return sni
	}
	return u.Hostname()
}

func fixPadding(b64 string) string {
	b64 = strings.ReplaceAll(strings.ReplaceAll(b64, "-", "+"), "_", "/")
	if pad := len(b64) % 4; pad > 0 {
		b64 += strings.Repeat("=", 4-pad)
	}
	return b64
}
