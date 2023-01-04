package main

import (
	"testing"

	"github.com/iamburbo/zacks-scraper/config"
	"github.com/iamburbo/zacks-scraper/util"
	"github.com/iamburbo/zacks-scraper/zacks"
)

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
	match := util.FindStringInBetween(left, right, left+str+right)
	if match != str {
		t.Fail()
	}
}

func TestLogin(t *testing.T) {
	client := util.GetTestClient()
	config, err := config.LoadConfigFile("config.yml")
	if err != nil {
		t.Fatal(err)
	}

	err = zacks.LogIn(client, config)
	if err != nil {
		t.Fatal(err)
	}
}
