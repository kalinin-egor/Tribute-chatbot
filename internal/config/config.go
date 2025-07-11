package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит все настройки приложения
type Config struct {
	TelegramBotToken    string
	TelegramAdminChatID int64
	LogLevel            string
	Port                int
	APIBaseURL          string
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config := &Config{
		TelegramBotToken:    getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramAdminChatID: getEnvAsInt64("TELEGRAM_ADMIN_CHAT_ID", 0),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		Port:                getEnvAsInt("PORT", 8080),
		APIBaseURL:          getEnv("API_BASE_URL", ""),
	}

	if config.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	if config.APIBaseURL == "" {
		return nil, fmt.Errorf("API_BASE_URL is required")
	}

	if config.TelegramAdminChatID == 0 {
		return nil, fmt.Errorf("TELEGRAM_ADMIN_CHAT_ID is required")
	}

	return config, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64 получает значение переменной окружения как int64 или возвращает значение по умолчанию
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
