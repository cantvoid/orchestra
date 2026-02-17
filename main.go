package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"orchestra/fetcher"
	"orchestra/parser"
	"orchestra/proxy"
	"os"
	"time"
)

func getBestProxy(subscriptionLink string) (string, error) {
	links, err := fetcher.GetLinks(subscriptionLink)

	if err != nil {
		return "", err
	}

	type result struct {
		link    string
		latency int
	}

	results := make(chan result, len(links))

	for _, link := range links {
		go func(l string) {
			latency := proxy.GetProxyLatency(l)
			results <- result{
				link:    l,
				latency: latency,
			}
		}(link)
	}

	var bestLink string
	minLatency := math.MaxInt

	for i := 0; i < len(links); i++ {
		res := <-results
		if res.latency != -1 && res.latency < minLatency {
			minLatency = res.latency
			bestLink = res.link
		}
	}

	if bestLink == "" {
		return "", fmt.Errorf("no valid proxies found")
	}

	return bestLink, nil
}

func main() {
	subscriptionLink := flag.String("link", "", "subscription link")
	flag.StringVar(subscriptionLink, "l", "", "subscription link")

	singboxPath := flag.String("singbox-path", "", "path to sing-box binary")
	flag.StringVar(singboxPath, "s", "", "path to sing-box binary")

	flag.Parse()

	if *subscriptionLink == "" || *singboxPath == "" {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "both --link (-l) and --singbox-path (-s) are required\n")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for {
		bestProxy, err := getBestProxy(*subscriptionLink)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get best proxy: %s\n", err)
			time.Sleep(3 * time.Second)
			continue
		}
		config, err := parser.ProxyToSingbox(bestProxy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't parse config: %s\n", err)
			time.Sleep(3 * time.Second)
			continue
		}
		fmt.Printf("starting sing-box with link %s\n", bestProxy)
		proc, err := proxy.StartTun(config, *singboxPath)

		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't start proxy: %s\n", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for {
			for i := 0; i < 3; i++ {
				_, err = client.Get("http://google.com/generate_204")
				if err == nil {
					break
				}
				time.Sleep(3 * time.Second)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "internet seems down, rotating\n")
				proc.Terminate()
				break
			} else {
				time.Sleep(10 * time.Second)
			}
		}
	}

}
