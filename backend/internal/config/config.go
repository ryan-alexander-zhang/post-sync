package config

import (
	"os"
	"strings"
	"strconv"
	"time"
)

type Config struct {
	AppEnv           string
	ServerAddr       string
	DBDriver         string
	DBDSN            string
	CORSAllowOrigins []string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	PublishConfig    PublishConfig
}

type PublishConfig struct {
	MaxParallelism int
	Timeout        time.Duration
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),
		DBDriver:         getEnv("DB_DRIVER", "sqlite"),
		DBDSN:            getEnv("DB_DSN", "./data/post-sync.db"),
		CORSAllowOrigins: getEnvAsCSV("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000"}),
		HTTPReadTimeout:  getEnvAsDuration("HTTP_READ_TIMEOUT_SECONDS", 10*time.Second),
		HTTPWriteTimeout: getEnvAsDuration("HTTP_WRITE_TIMEOUT_SECONDS", 30*time.Second),
		PublishConfig: PublishConfig{
			MaxParallelism: getEnvAsInt("PUBLISH_MAX_PARALLELISM", 5),
			Timeout:        getEnvAsDuration("PUBLISH_TIMEOUT_SECONDS", 20*time.Second),
		},
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

func getEnvAsInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return value
}

func getEnvAsCSV(key string, fallback []string) []string {
	raw := os.Getenv(key)
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	items := strings.Split(raw, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}

	return values
}
