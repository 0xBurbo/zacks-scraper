package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/stockscreener"
	"github.com/iamburbo/zacks-scraper/zacks"
	"golang.org/x/net/publicsuffix"
)

func main() {
	// Load config
	configPath, err := config.ParseConfigPathFromArgs()
	if err != nil {
		log.Fatalf("Error parsing command line args: %v", err)
	}

	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		log.Fatalf("Error loading config file: %v", err)
	}

	// Setup http client
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{
		Jar: jar,
	}

	// Retreive logged in session
	err = zacks.LogIn(client, cfg)
	if err != nil {
		log.Fatalf("Error while logging in: %v", err)
	}

	// Execute each job from the config, retrying if necessary
	for _, job := range cfg.Jobs {
		retry := 0
		for retry < cfg.MaxRetries {
			var err error
			err = nil
			switch job.JobType {
			case "stock_screener":
				err = stockscreener.RunStockScreener(&job, client)
			}

			if err != nil {
				retry++
				time.Sleep(time.Duration(cfg.DelayBetweenRetries) * time.Millisecond)
			} else {
				break
			}
		}

	}
}
