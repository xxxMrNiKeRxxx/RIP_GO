package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	host := os.Getenv("APP_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("APP_PORT")
	if portStr == "" {
		portStr = "8080"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %v", err)
	}

	return &Config{
		ServiceHost: host,
		ServicePort: port,
	}, nil
}
