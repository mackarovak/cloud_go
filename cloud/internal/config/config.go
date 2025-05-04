package config

import (
	"os"
	"time"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int       `yaml:"port"`
	Backends []Backend `yaml:"backends"`

	RateLimiter struct {
		DefaultCapacity  int                        `yaml:"default_capacity"`
		DefaultRate      int                        `yaml:"default_rate"`
		RefillInterval   time.Duration             `yaml:"refill_interval"`
		ClientSpecific   map[string]ClientConfig   `yaml:"client_specific"`
	} `yaml:"rate_limiter"`
}

type Backend struct {
	URL     string `yaml:"url"`
	Healthy bool   `yaml:"healthy"`
}

type ClientConfig struct {
	Capacity int `yaml:"capacity"`
	Rate     int `yaml:"rate"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}