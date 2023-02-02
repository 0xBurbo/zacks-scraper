package util

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Creates a test client that runs request through Charles proxy
// NOTE: All tests will fail unless Charles proxy is running on port 8888
func GetTestClient() *http.Client {
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

func FindStringInBetween(left, right, str string) string {
	// Get substring between two strings.
	posFirst := strings.Index(str, left)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(str, right)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(left)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return str[posFirstAdjusted:posLast]
}
