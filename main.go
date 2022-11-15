package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/publicsuffix"
)

func main() {
	// Retrieve settings
	outDir, config, err := ParseArgsAndReadInputFile()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Setup http client
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// Send entire request flow, retry if errors occur
	retry := 0
	var data [][]string
	for retry < config.MaxRetries {
		data, err = requestFlow(client, config)
		if err != nil {
			retry++
			log.Println(err)

			// Exit if max retries has been exceeded
			if retry >= config.MaxRetries {
				log.Fatalf("Request flow failed after %v retries", retry)
			}

			// Wait retry delay and try again
			time.Sleep(time.Duration(config.DelayBetweenRetries) * time.Millisecond)
			continue
		}
		break
	}

	// Write new output file to directory
	fileName := time.Now().Format("2006-01-02_15-04-05") + ".csv"
	f, err := os.Create(filepath.Join(outDir, fileName))
	if err != nil {
		log.Fatalf("error creating output file: %v", err)
	}
	w := csv.NewWriter(f)
	defer w.Flush()

	for _, record := range data {
		if err = w.Write(record); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}
}

func requestFlow(client *http.Client, config *zacksScraperInput) ([][]string, error) {
	prefix := "an error occured while"

	// HTTP Request flow for retreiving stock screener data

	// Retreive session cookie using login credentials
	err := LogIn(client, config)
	if err != nil {
		return nil, fmt.Errorf("%v logging in: %w", prefix, err)
	}

	parsedStockScreenerPage, err := GetStockScreenerPage(client)
	if err != nil {
		return nil, fmt.Errorf("%v fetching stock screener page: %w", prefix, err)
	}

	err = GetScreenerFromApi(client, parsedStockScreenerPage)
	if err != nil {
		return nil, fmt.Errorf("%v fetching screener API page: %w", prefix, err)
	}

	err = ResetParam(client)
	if err != nil {
		return nil, fmt.Errorf("%v resetting query params: %w", prefix, err)
	}

	err = QueryScreenerApi(client, config)
	if err != nil {
		return nil, fmt.Errorf("%v sending query: %w", prefix, err)
	}

	data, err := DownloadData(client, parsedStockScreenerPage)
	if err != nil {
		return nil, fmt.Errorf("%v resetting query params: %w", prefix, err)
	}
	return data, nil
}
