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

func getBestProxy(subscriptionLink string, timeoutTime time.Duration) (string, error) {
	links, err := fetcher.GetLinks(subscriptionLink, timeoutTime)

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

	waitTime := flag.Duration("wait", 5*time.Second, "how much time to wait for sing-box to start (use this when sing-box can't read the config fast enough)")

	testLink := flag.String("testwith", "https://google.com/generate_204", "the uri for testing internet connection")

	timeoutTime := flag.Duration("timeout", 30*time.Second, "how much to wait for timeout during HTTP requests")
	pollTime := flag.Duration("poll", 10*time.Second, "how fast to poll connection")

	flag.Parse()

	if *subscriptionLink == "" || *singboxPath == "" {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "both --link (-l) and --singbox-path (-s) are required\n")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for {
		bestProxy, err := getBestProxy(*subscriptionLink, *timeoutTime)
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
		proc, err := proxy.StartTun(config, *singboxPath, *waitTime)

		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't start proxy: %s\n", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for {
			for i := 0; i < 3; i++ {
				_, err = client.Get(*testLink)
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
				time.Sleep(*pollTime)
			}
		}
	}

}
