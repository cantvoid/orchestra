package proxy

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"net"
	"net/url"

	gopsnet "github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	"encoding/json"
)

func killPortHogs(port int) {
	connections, _ := gopsnet.Connections("tcp")
	currentPid := int32(os.Getpid())

	for _, conn := range connections {
		if conn.Laddr.Port == uint32(port) && conn.Pid != currentPid {
			if process, err := process.NewProcess(conn.Pid); err == nil {
				process.Kill()
			}
		}
	}
}
func GetProxyLatency(uri string) int {
	parsed, err := url.Parse(uri)
	if err != nil {
		return -1
	}
	host := parsed.Hostname()
	port := parsed.Port()

	start := time.Now().UnixMilli()
	var address string
	if port != "" {
		address = host + ":" + port
	} else {
		address = host + ":443"
	}
	connection, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return -1
	}
	defer connection.Close()
	stop := time.Now().UnixMilli()
	elapsedTime := stop - start

	return int(elapsedTime)
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

	cmd := exec.Command(singboxPath, "run", "-c", tmpPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}
	time.Sleep(waitTime)
	os.Remove(tmpPath)

	proc, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return nil, err
	}
	return proc, nil
}
