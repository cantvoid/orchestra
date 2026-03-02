package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetBestProxy(t *testing.T) {
	const encodedLinksBad = "dm1lc3M6Ly9mYWtldm1lc3MNCnZsZXNzOi8vbm90YXJlYWxhZGRyZXNzLmZha2U="
	const encodedLinksGood = "dmxlc3M6Ly9leGFtcGxlLmNvbQp0cm9qYW46Ly90ZXN0QGV4YW1wbGUuY29tCnZtZXNzOi8vZXlKMklqb2lNaUlzSW5Ceklqb2lUWGtnVTJWeWRtVnlJaXdpWVdSa0lqb2lNVEEwTGpFNExqSTJMakV5TUNJc0luQnZjblFpT2lJME5ETWlMQ0pwWkNJNkltRmhZV0ZoWVdGaExXSmlZbUl0WTJOall5MWtaR1JrTFdWbFpXVmxaV1ZsWldVaUxDSmhhV1FpT2lJd0lpd2libVYwSWpvaWQzTWlMQ0owZVhCbElqb2libTl1WlNJc0ltaHZjM1FpT2lKbGVHRnRjR3hsTG1OdmJTSXNJbkJoZEdnaU9pSXZjbUY1SWl3aWRHeHpJam9pZEd4ekluMD0="
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bad":
			fmt.Fprint(w, encodedLinksBad)
		case "/empty":
			w.WriteHeader(http.StatusOK)
		default:
			fmt.Fprint(w, encodedLinksGood)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Success", server.URL, false},
		{"Bad data", server.URL + "/bad", true},
		{"Empty response", server.URL + "/empty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getBestProxy(tt.url, 60*time.Second)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}

// func FetchSingbox() (string, error) {
// 	path, err := exec.LookPath("sing-box.exe")
// 	if err == nil {
// 		return path, nil
// 	}

// 	fmt.Println("sing-box not found in PATH, attempting to install...")
// 	installCmd := exec.Command("go", "install", "github.com/sagernet/sing-box/cmd/sing-box@latest")
// 	if err := installCmd.Run(); err != nil {
// 		return "", fmt.Errorf("failed to install sing-box: %w", err)
// 	}
// 	time.Sleep(2 * time.Second)
// 	return exec.LookPath("sing-box.exe")
// }

func TestMain(t *testing.T) {
	const encodedLinksGood = "dmxlc3M6Ly9leGFtcGxlLmNvbQp0cm9qYW46Ly90ZXN0QGV4YW1wbGUuY29tCnZtZXNzOi8vZXlKMklqb2lNaUlzSW5Ceklqb2lUWGtnVTJWeWRtVnlJaXdpWVdSa0lqb2lNVEEwTGpFNExqSTJMakV5TUNJc0luQnZjblFpT2lJME5ETWlMQ0pwWkNJNkltRmhZV0ZoWVdGaExXSmlZbUl0WTJOall5MWtaR1JrTFdWbFpXVmxaV1ZsWldVaUxDSmhhV1FpT2lJd0lpd2libVYwSWpvaWQzTWlMQ0owZVhCbElqb2libTl1WlNJc0ltaHZjM1FpT2lKbGVHRnRjR3hsTG1OdmJTSXNJbkJoZEdnaU9pSXZjbUY1SWl3aWRHeHpJam9pZEd4ekluMD0="
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, encodedLinksGood)
	}))
	defer server.Close()

	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		watnedMsg string
	}{
		{"Invalid sing-box path", []string{"orchestra", "-l", server.URL, "-s", "invalidPath"}, true, "exec: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exit := os.Exit
			exit = func(code int) {}
			exit(1)
			r, w, _ := os.Pipe()
			oldStderr := os.Stderr
			os.Stderr = w
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			go main()
			outChan := make(chan string)
			go func() {
				scanner := bufio.NewScanner(r)
				if scanner.Scan() {
					outChan <- scanner.Text()
				}
			}()
			select {
			case msg := <-outChan:
				if !strings.Contains(msg, tt.watnedMsg) {
					t.Errorf("unexpected output: %s", msg)
				}
				w.Close()
				os.Stderr = oldStderr
			case <-time.After(2 * time.Second):
				t.Fatal("test timed out: main() didn't print to stderr")
			}
		})
	}
}
