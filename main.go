package main

import (
	"tribute-chatbot/internal/bot"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	godotenv.Load()

	// Инициализируем логгер
	logg := logger.New()

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		logg.Fatal("Failed to load configuration", err)
	}

	// Создаем и запускаем бота
	botInstance, err := bot.NewBot(cfg)
	if err != nil {
		logg.Fatal("Failed to create bot", err)
	}

	// Запускаем бота
	botInstance.Start()
}
