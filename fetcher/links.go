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
	} else {
		method = "xrayjson"
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
	}
	return decodedLinks, nil
}
