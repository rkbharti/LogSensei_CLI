package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Pattern struct {
	Name    string `yaml:"name"`
	Keyword string `yaml:"keyword"`
}

type Config struct {
	Patterns []Pattern `yaml:"patterns"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}