package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	ServerPort      string
	DBPath          string
	LogLevel        slog.Level
	NagerAPIBaseURL string
}

func Load() *Config {
	return &Config{
		ServerPort:      getEnv("SERVER_PORT", "9119"),
		DBPath:          getEnv("DB_PATH", "citynext.db"),
		LogLevel:        parseLogLevel(getEnv("LOG_LEVEL", "info")),
		NagerAPIBaseURL: getEnv("NAGER_API_BASE_URL", "https://date.nager.at/api/v3"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
