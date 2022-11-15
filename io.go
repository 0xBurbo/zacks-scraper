package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type zacksScraperInput struct {
	Username            string                   `json:"username"`
	Password            string                   `json:"password"`
	MaxRetries          int                      `json:"maxRetries"`
	DelayBetweenRetries int                      `json:"delayBetweenRetries"`
	Queries             []map[string]interface{} `json:"queries"`
}

func ParseArgsAndReadInputFile() (string, *zacksScraperInput, error) {
	args := os.Args

	var inputFilepath, outputFilepath string

	for _, v := range args {
		if strings.Contains(v, "--config=") {
			split := strings.Split(v, "=")
			if len(split) == 2 {
				inputFilepath = split[1]
			} else {
				log.Fatalf("Invalid arg: %v", v)
			}
		} else if strings.Contains(v, "--outDir=") {
			split := strings.Split(v, "=")
			if len(split) == 2 {
				outputFilepath = split[1]
			} else {
				log.Fatalf("Invalid arg: %v", v)
			}
		}
	}

	config, err := readInput(inputFilepath)
	return outputFilepath, config, err
}

func readInput(path string) (*zacksScraperInput, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	jsonFile, err := os.Open(p)

	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	input := &zacksScraperInput{}
	json.Unmarshal(jsonBytes, input)
	return input, nil
}
