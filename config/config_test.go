package config

import (
	"log"
	"testing"
)

func TestParseConfig(t *testing.T) {
	path := "../config.yml"
	config, err := LoadConfigFile(path)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(*config)
}
