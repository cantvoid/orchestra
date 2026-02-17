package fetcher

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetLinks(subscriptionLink string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", subscriptionLink, nil)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"User-Agent":      "Happ/9.9.9/Windows",
		"X-App-Version":   "9.9.9",
		"X-Device-Locale": "EN",
		"X-Device-Os":     "Windows",
		"X-Device-Model":  "orchestra",
		"X-Hwid":          "orchestra",
		"X-Ver-Os":        "i forgot sorry",
		"Connection":      "Keep-Alive",
		"Accept-Language": "*",
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d from %s", resp.StatusCode, subscriptionLink)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}
	decodedLinks := strings.Split(string(decoded), "\n")
	return decodedLinks, nil
}
