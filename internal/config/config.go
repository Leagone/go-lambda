package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ApiUrl                url.URL
	HandlerRequestTimeOut time.Duration
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

	timeoutStr := os.Getenv("HANDLER_PROCESS_REQUEST_TIMEOUT")
	if timeoutStr == "" {
		return nil, fmt.Errorf("HANDLER_PROCESS_REQUEST_TIMEOUT is missing")
	}

	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid RECORD_TIMEOUT_SECONDS: %w", err)
	}

	timeoutDuration := time.Duration(timeoutSeconds) * time.Second

	return &Config{
		ApiUrl:                *url,
		HandlerRequestTimeOut: timeoutDuration}, nil

}
