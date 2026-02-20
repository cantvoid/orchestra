package proxy

import (
	"fmt"
	"orchestra/parser"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"net"
	"net/url"

	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"encoding/json"
)

func killPortHogs(port int) error {
	connections, err := gopsnet.Connections("tcp")
	if err != nil {
		return err
	}
	currentPid := int32(os.Getpid())

	for _, conn := range connections {
		if conn.Laddr.Port == uint32(port) && conn.Pid != currentPid {
			if p, err := process.NewProcess(conn.Pid); err == nil {
				if err := p.Kill(); err != nil {
					fmt.Fprintf(os.Stderr, "failed to kill pid %d: %v\n", conn.Pid, err)
				}
			}
		}
	}
	return nil
}
func GetProxyLatency(uri string) (int, error) {
	var host string
	var port string
	var ok bool
	var portint int
	if strings.HasPrefix(uri, "vmess://") {
		parsed, err := parser.VmessToSingbox(uri)
		if err != nil {
			return -1, fmt.Errorf("error while trying to parse vmess link '%s': %v\n", uri, err)
		}
		host, ok = parsed["server"].(string)
		if !ok {
			return -1, fmt.Errorf("expected host to be string, is %T\n", parsed["server"])
		}

		portint, ok = parsed["server_port"].(int)
		if !ok {
			return -1, fmt.Errorf("expected port to be an int, is %T\n", parsed["server_port"])
		}
		port = strconv.Itoa(portint)
	} else if strings.HasPrefix(uri, "trojan://") {
		parsed, err := parser.TrojanToSingbox(uri)
		if err != nil {
			return -1, fmt.Errorf("error while trying to parse trojan link '%s': %v\n", uri, err)
		}

		host, ok = parsed["server"].(string)
		if !ok {
			return -1, fmt.Errorf("expected host to be an int, is %T\n", parsed["server"])
		}

		portint, ok = parsed["server_port"].(int)
		if !ok {
			return -1, fmt.Errorf("expected port to be an int, is %T\n", parsed["server_port"])
		}
		port = strconv.Itoa(portint)
	} else {
		parsed, err := url.Parse(uri)
		if err != nil {
			return -1, fmt.Errorf("error while trying to parse link '%s': %v\n", uri, err)
		}
		host = parsed.Hostname()
		port = parsed.Port()
	}

	start := time.Now().UnixMilli()
	var address string
	if port != "" {
		address = host + ":" + port
	} else {
		address = host + ":443"
	}
	connection, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return -1, err
	}
	defer connection.Close()
	stop := time.Now().UnixMilli()
	elapsedTime := stop - start

	return int(elapsedTime), nil
}

func StartTun(config map[string]interface{}, singboxPath string, waitTime time.Duration) (*process.Process, error) {
	killPortHogs(10808)

	tmpFile, err := os.CreateTemp("", "dynamic_config.json")
	if err != nil {
		return nil, err
	}
	tmpPath, _ := filepath.Abs(tmpFile.Name())

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return nil, err
	}
	tmpFile.Close()

	cmd := exec.Command(singboxPath, "run", "-c", tmpPath, "--disable-color")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}
	time.Sleep(waitTime)
	os.Remove(tmpPath)

	exists, err := process.PidExists(int32(cmd.Process.Pid))
	if err != nil {
		return nil, fmt.Errorf("failed to check process: %w\n", err)
	}
	if !exists {
		return nil, fmt.Errorf("sing-box died during startup\n")
	}

	proc, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}

	return proc, nil
}
