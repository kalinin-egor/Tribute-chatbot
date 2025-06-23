package main

import (
	"log"

	"tribute-chatbot/internal/bot"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load("config.env"); err != nil {
		log.Printf("Warning: config.env file not found, using system environment variables")
	}

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
