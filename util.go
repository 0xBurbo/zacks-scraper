package main

import (
	"errors"
	"net/url"
	"strings"
)

func findStringInBetween(left, right, str string) string {
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

type parsedStockScreenerHomePage struct {
	CKey string
}

func parseStockScreenerHomePage(body string) (*parsedStockScreenerHomePage, error) {
	substr := findStringInBetween(`<iframe style="" title="Stock Screener " id="screenerContent" src="`, `" scrolling="yes" allowfullscreen></iframe>`, body)
	parsed, err := url.Parse(substr)
	if err != nil {
		return nil, err
	}

	ckey := parsed.Query().Get("c_key")
	if ckey == "" {
		return nil, errors.New("no c_key found")
	}
	return &parsedStockScreenerHomePage{
		CKey: ckey,
	}, nil
}
