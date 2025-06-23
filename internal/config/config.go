package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит все настройки приложения
type Config struct {
	TelegramBotToken string
	LogLevel         string
	Port             int
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	config := &Config{
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
		Port:             getEnvAsInt("PORT", 8080),
	}

	if config.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
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
