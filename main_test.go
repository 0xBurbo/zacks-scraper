package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"golang.org/x/net/publicsuffix"
)

func getTestClient() *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}

	proxyUrl, err := url.Parse("http://localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	return &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
}

func TestReadInputJson(t *testing.T) {
	_, err := readInput("example.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFindStringInBetween(t *testing.T) {
	left := "aba"
	right := "bcb"
	str := "test"
	match := findStringInBetween(left, right, left+str+right)
	if match != str {
		t.Fail()
	}
}

func TestLogin(t *testing.T) {
	client := getTestClient()
	config, err := readInput("config.json")
	if err != nil {
		t.Fatal(err)
	}

	err = LogIn(client, config)
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryScreenerApi(t *testing.T) {
	client := getTestClient()
	config, err := readInput("config.json")
	if err != nil {
		t.Fatal(err)
	}

	err = LogIn(client, config)
	if err != nil {
		t.Fatal(err)
	}

	err = QueryScreenerApi(client, config)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDownloadData(t *testing.T) {
	client := getTestClient()
	config, err := readInput("config.json")
	if err != nil {
		t.Fatal(err)
	}

	err = LogIn(client, config)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := GetStockScreenerPage(client)
	if err != nil {
		t.Fatal(err)
	}

	err = GetScreenerFromApi(client, parsed)
	if err != nil {
		t.Fatal(err)
	}

	err = ResetParam(client)
	if err != nil {
		t.Fatal(err)
	}

	err = QueryScreenerApi(client, config)
	if err != nil {
		t.Fatal(err)
	}

	_, err = DownloadData(client, parsed)
	if err != nil {
		t.Fatal(err)
	}
}
