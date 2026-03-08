package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv           string
	ServerAddr       string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),
		HTTPReadTimeout:  getEnvAsDuration("HTTP_READ_TIMEOUT_SECONDS", 10*time.Second),
		HTTPWriteTimeout: getEnvAsDuration("HTTP_WRITE_TIMEOUT_SECONDS", 30*time.Second),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
