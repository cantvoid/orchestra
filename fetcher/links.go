package fetcher

import (
	"encoding/base64"
	"fmt"
	"orchestra/parser"
	"strings"
)

func BodyToLink(body []byte) ([]string, error) {
	var method string
	_, err := base64.StdEncoding.DecodeString(string(body))
	if err == nil {
		method = "base64"
	} else if strings.HasPrefix(string(body), "[") && strings.HasSuffix(string(body), "]") {
		method = "xrayjson"
	} else if strings.Contains(string(body), "\n") && strings.Contains(string(body), "://") { //hacky
		method = "plain"
	}

	var decodedLinks []string
	switch method {
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			return nil, err
		}
		decodedLinks = strings.Split(string(decoded), "\n")

	case "xrayjson":
		decodedLinks, err = parser.ConvertToLinks(body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subscription data: %w", err)
		}
	case "plain":
		decodedLinks = strings.Split(string(body), "\n")
	default:
		return nil, fmt.Errorf("unkown subscription data format")
	}
	return decodedLinks, nil
}
