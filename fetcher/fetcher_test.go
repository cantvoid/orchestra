package fetcher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetLinks(t *testing.T) {
	const encodedLinks = "dm1lc3M6Ly9leUpoWkdRaU9pSmxlR0Z0Y0d4bExtTnZiU0lzSW1GcFpDSTZJakFpTENKb2IzTjBJam9pSWl3aWFXUWlPaUptWVd0bExXbGtMVEV5TXpRMU5qYzRPVEFpTENKdVpYUWlPaUowWTNBaUxDSndZWFJvSWpvaUlpd2ljRzl5ZENJNklqUTBNeUlzSW5Ceklqb2lWR1Z6ZEMxV2JXVnpjeUlzSW5Sc2N5STZJaUlzSW5SNWNHVWlPaUp1YjI1bElpd2lkaUk2SWpJaWZRPT0Kdmxlc3M6Ly9mYWtlLWlkLWFiY2RlZi0xMjM0NTZAZXhhbXBsZS5jb206NDQzP2VuY3J5cHRpb249bm9uZSZzZWN1cml0eT10bHMmc25pPWV4YW1wbGUuY29tJmZwPWNocm9tZSZwYms9ZmFrZXB1YmxpY2tleSZzaWQ9ZmFrZXNpZCZ0eXBlPXRjcCZoZWFkZXJUeXBlPW5vbmUjVGVzdC1WTEVTUwp0cm9qYW46Ly9mYWtlcGFzc3dvcmQxMjNBZXhhbXBsZS5jb206NDQzP3NlY3VyaXR5PXRscyZzbmk9ZXhhbXBsZS5jb20mdHlwZT10Y3AmaGVhZGVyVHlwZT1ub25lI1Rlc3QtVHJvamFuCnZtZXNzOi8vZXlKaFpHUWlPaUpsZUdGdGNHeGxMbU52YlNJc0ltRnBaQ0k2SWpBaUxDSm9iM04wSWpvaVpYaGhiWEJzWlM1amIyMGlMQ0pwWkNJNkltWmhhMlV0ZG5OcFpDMWhZbU5rWlNJc0ltNWxkQ0k2SW5keklpd2ljR0YwYUNJNklpOW1ZV3RsTDNkeklpd2ljRzl5ZENJNklqUTBNeUlzSW5Ceklqb2lWR1Z6ZEMxV2JXVnpjeTFYVXlJc0luUnNjeUk2SW5Sc2N5SXNJblI1Y0dVaU9pSnViMjVsSWl3aWRpSTZJaklpZlE9PQp2bGVzczovL2Zha2UtZ3JwYy1pZEBleGFtcGxlLmNvbTo0NDM/ZW5jcnlwdGlvbj1ub25lJnNlY3VyaXR5PXRscyZ0eXBlPWdycGMmc2VydmljZU5hbWU9ZmFrZS5zZXJ2aWNlJnNuaT1leGFtcGxlLmNvbSNUZXN0LVZMRVNTLWdSUEM="

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/content-length":
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Short body"))
			hj, _ := w.(http.Hijacker)
			conn, _, _ := hj.Hijack()
			conn.Close()
		case "/malformed":
			fmt.Fprint(w, "malformed data")
		case "/empty":
			w.WriteHeader(http.StatusOK)
		case "/403":
			w.WriteHeader(http.StatusForbidden)
		default:
			fmt.Fprint(w, encodedLinks)
		}
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		wantErr bool
		count   int
	}{
		{"Success", server.URL, false, 5},
		{"Connection Loss", server.URL + "/content-length", true, 0},
		{"Bad Base64", server.URL + "/malformed", true, 0},
		{"Empty Body", server.URL + "/empty", true, 0},
		{"Forbidden 403", server.URL + "/403", true, 0},
		{"Malformed Link", "!!://bad-url-\x7f", true, 0},
		{"Invalid Port", "http://localhost:0", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links, err := GetLinks(tt.url, 5*time.Second)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLinks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(links) != tt.count {
				t.Errorf("expected %d links, got %d", tt.count, len(links))
			}
		})
	}
}
