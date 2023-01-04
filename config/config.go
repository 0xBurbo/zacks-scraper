package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Username            string      `yaml:"username"`
	Password            string      `yaml:"password"`
	MaxRetries          int         `yaml:"maxRetries"`
	DelayBetweenRetries int         `yaml:"delayBetweenRetries"`
	Jobs                []ScrapeJob `yaml:"jobs"`
}

type ScrapeJob struct {
	JobType    string                   `yaml:"jobType"`
	OutDir     string                   `yaml:"outDir"`
	Parameters []map[string]interface{} `yaml:"parameters"`
}

func ParseConfigPathFromArgs() (string, error) {
	args := os.Args

	var configPath string

	for _, v := range args {
		if strings.Contains(v, "--config=") {
			split := strings.Split(v, "=")
			if len(split) == 2 {
				configPath = split[1]
			} else {
				return "", fmt.Errorf("invalid arg: %v", v)
			}
		}
	}

	return configPath, nil
}

func LoadConfigFile(path string) (*Config, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	configFile, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	configBytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
