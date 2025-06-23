package main

import (
	"tribute-chatbot/internal/bot"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env файла.
	// Ошибки игнорируются, если файл не найден,
	// так как переменные могут быть установлены системно.
	godotenv.Load() // по умолчанию загружает .env

	// Инициализируем логгер
	logger := logger.New()

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
	}

	// Создаем и запускаем бота
	botInstance, err := bot.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create bot", err)
	}

	logger.Info("Starting Telegram bot...")
	if err := botInstance.Start(); err != nil {
		logger.Fatal("Bot stopped with error", err)
	}
}
