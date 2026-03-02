package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func ConvertToLinks(data []byte) ([]string, error) {
	var configs []map[string]interface{}
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var links []string
	for _, cfg := range configs {
		remarks, _ := cfg["remarks"].(string)

		outbounds, _ := cfg["outbounds"].([]interface{})
		var outbound map[string]interface{}
		for _, ob := range outbounds {
			if m, ok := ob.(map[string]interface{}); ok {
				proto, _ := m["protocol"].(string)
				if proto != "freedom" && proto != "blackhole" && proto != "dns" {
					outbound = m
					break
				}
			}
		}
		if outbound == nil {
			continue
		}

		protocol, _ := outbound["protocol"].(string)
		settings, _ := outbound["settings"].(map[string]interface{})
		stream, _ := outbound["streamSettings"].(map[string]interface{})

		var link string
		var err error

		switch protocol {
		case "vless":
			link, err = buildVlessLink(settings, stream)
		case "vmess":
			link, err = buildVmessLink(settings, stream)
		case "trojan":
			link, err = buildTrojanLink(settings, stream)
		default:
			fmt.Fprintf(os.Stderr, "unsupported protocol: %s\n", protocol)
			continue
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid outbound, %v\n", err)
			continue
		}

		if remarks != "" {
			link += "#" + url.PathEscape(remarks)
		}
		links = append(links, link)
	}
	return links, nil
}

func buildVlessLink(settings, stream map[string]interface{}) (string, error) {
	vnext, _ := settings["vnext"].([]interface{})
	if len(vnext) == 0 {
		return "", fmt.Errorf("missing vnext")
	}
	srv := vnext[0].(map[string]interface{})
	addr, _ := srv["address"].(string)
	port, _ := srv["port"].(float64)
	users, _ := srv["users"].([]interface{})
	if len(users) == 0 {
		return "", fmt.Errorf("missing users")
	}
	user := users[0].(map[string]interface{})
	id, _ := user["id"].(string)

	link := fmt.Sprintf("vless://%s@%s:%d", id, addr, int(port))

	params := url.Values{}
	if enc, _ := user["encryption"].(string); enc != "" && enc != "none" {
		params.Set("encryption", enc)
	}
	if flow, _ := user["flow"].(string); flow != "" {
		params.Set("flow", flow)
	}
	appendStreamParams(&params, stream)

	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	return link, nil
}

func buildVmessLink(settings, stream map[string]interface{}) (string, error) {
	vnext, _ := settings["vnext"].([]interface{})
	if len(vnext) == 0 {
		return "", fmt.Errorf("missing vnext")
	}
	srv := vnext[0].(map[string]interface{})
	addr, _ := srv["address"].(string)
	port, _ := srv["port"].(float64)
	users, _ := srv["users"].([]interface{})
	if len(users) == 0 {
		return "", fmt.Errorf("missing users")
	}
	user := users[0].(map[string]interface{})
	id, _ := user["id"].(string)
	aid, _ := user["alterId"].(float64)
	sec, _ := user["security"].(string)
	if sec == "" {
		sec = "auto"
	}

	vmessObj := map[string]interface{}{
		"v":    "2",
		"ps":   "",
		"add":  addr,
		"port": int(port),
		"id":   id,
		"aid":  int(aid),
		"net":  "tcp",
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "",
		"sni":  "",
	}
	mergeStreamIntoVMess(vmessObj, stream)

	data, err := json.Marshal(vmessObj)
	if err != nil {
		return "", err
	}
	return "vmess://" + base64.StdEncoding.EncodeToString(data), nil
}

func buildTrojanLink(settings, stream map[string]interface{}) (string, error) {
	servers, _ := settings["servers"].([]interface{})
	if len(servers) == 0 {
		return "", fmt.Errorf("missing servers")
	}
	srv := servers[0].(map[string]interface{})
	addr, _ := srv["address"].(string)
	port, _ := srv["port"].(float64)
	pass, _ := srv["password"].(string)

	link := fmt.Sprintf("trojan://%s@%s:%d", pass, addr, int(port))

	params := url.Values{}
	appendStreamParams(&params, stream)

	if len(params) > 0 {
		link += "?" + params.Encode()
	}
	return link, nil
}

func appendStreamParams(params *url.Values, stream map[string]interface{}) {
	if stream == nil {
		return
	}
	if net, _ := stream["network"].(string); net != "" && net != "tcp" {
		params.Set("type", net)
	}
	if sec, _ := stream["security"].(string); sec != "" {
		params.Set("security", sec)

		if reality, ok := stream["realitySettings"].(map[string]interface{}); ok && sec == "reality" {
			if fp, _ := reality["fingerprint"].(string); fp != "" {
				params.Set("fp", fp)
			}
			if pbk, _ := reality["publicKey"].(string); pbk != "" {
				params.Set("pbk", pbk)
			}
			if sid, _ := reality["shortId"].(string); sid != "" {
				params.Set("sid", sid)
			}
			if sn, _ := reality["serverName"].(string); sn != "" {
				params.Set("sni", sn)
			}
		}
		if tls, ok := stream["tlsSettings"].(map[string]interface{}); ok && (sec == "tls" || sec == "xtls") {
			if sn, _ := tls["serverName"].(string); sn != "" {
				params.Set("sni", sn)
			}
			if alpn, _ := tls["alpn"].([]interface{}); len(alpn) > 0 {
				strs := make([]string, len(alpn))
				for i, v := range alpn {
					strs[i], _ = v.(string)
				}
				params.Set("alpn", strings.Join(strs, ","))
			}
			if fp, _ := tls["fingerprint"].(string); fp != "" {
				params.Set("fp", fp)
			}
		}
	}

	switch net, _ := stream["network"].(string); net {
	case "ws":
		if ws, ok := stream["wsSettings"].(map[string]interface{}); ok {
			if path, _ := ws["path"].(string); path != "" {
				params.Set("path", path)
			}
			if headers, ok := ws["headers"].(map[string]interface{}); ok {
				if host, _ := headers["Host"].(string); host != "" {
					params.Set("host", host)
				}
			}
		}
	case "h2", "http":
		if h2, ok := stream["httpSettings"].(map[string]interface{}); ok {
			if path, _ := h2["path"].(string); path != "" {
				params.Set("path", path)
			}
			if hosts, _ := h2["host"].([]interface{}); len(hosts) > 0 {
				if first, _ := hosts[0].(string); first != "" {
					params.Set("host", first)
				}
			}
		}
	case "grpc":
		if grpc, ok := stream["grpcSettings"].(map[string]interface{}); ok {
			if sn, _ := grpc["serviceName"].(string); sn != "" {
				params.Set("serviceName", sn)
			}
			if mode, _ := grpc["mode"].(string); mode != "" {
				params.Set("mode", mode)
			}
		}
	}
}

func mergeStreamIntoVMess(vmessObj map[string]interface{}, stream map[string]interface{}) {
	if stream == nil {
		return
	}
	if net, _ := stream["network"].(string); net != "" {
		vmessObj["net"] = net
	}
	if sec, _ := stream["security"].(string); sec == "tls" || sec == "xtls" || sec == "reality" {
		vmessObj["tls"] = sec
	}
	switch net, _ := stream["network"].(string); net {
	case "ws":
		if ws, ok := stream["wsSettings"].(map[string]interface{}); ok {
			if path, _ := ws["path"].(string); path != "" {
				vmessObj["path"] = path
			}
			if headers, ok := ws["headers"].(map[string]interface{}); ok {
				if host, _ := headers["Host"].(string); host != "" {
					vmessObj["host"] = host
				}
			}
		}
	case "h2", "http":
		if h2, ok := stream["httpSettings"].(map[string]interface{}); ok {
			if path, _ := h2["path"].(string); path != "" {
				vmessObj["path"] = path
			}
			if hosts, _ := h2["host"].([]interface{}); len(hosts) > 0 {
				if first, _ := hosts[0].(string); first != "" {
					vmessObj["host"] = first
				}
			}
		}
	}
	// SNI
	if tls, ok := stream["tlsSettings"].(map[string]interface{}); ok {
		if sn, _ := tls["serverName"].(string); sn != "" {
			vmessObj["sni"] = sn
		}
	}
	if reality, ok := stream["realitySettings"].(map[string]interface{}); ok {
		if sn, _ := reality["serverName"].(string); sn != "" {
			vmessObj["sni"] = sn
		}
	}
}
