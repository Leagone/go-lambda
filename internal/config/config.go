package config

import (
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	ApiUrl url.URL
}

func NewConfig() (*Config, error) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("missing API_BASE_URL env variable")
	}
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url for API_BASE_URL: %v", err)
	}

	return &Config{ApiUrl: *url}, nil

}
