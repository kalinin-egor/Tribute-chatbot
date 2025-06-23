package bot

import (
	"fmt"

	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/handlers"
	"tribute-chatbot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет Telegram бота
type Bot struct {
	api      *tgbotapi.BotAPI
	config   *config.Config
	logger   logger.Logger
	handlers *handlers.Handlers
}

// New создает новый экземпляр бота
func New(cfg *config.Config, log logger.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	// Создаем обработчики
	handlers := handlers.New(log)

	bot := &Bot{
		api:      api,
		config:   cfg,
		logger:   log,
		handlers: handlers,
	}

	// Настраиваем режим отладки
	api.Debug = cfg.LogLevel == "debug"

	log.Info("Bot authorized on account", api.Self.UserName)
	return bot, nil
}

// Start запускает бота
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.logger.Info("Bot started, waiting for messages...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Логируем входящие сообщения
		b.logger.Debug("Received message:", update.Message.Text, "from:", update.Message.From.UserName)

		// Обрабатываем сообщение
		if err := b.handleMessage(update.Message); err != nil {
			b.logger.Error("Error handling message:", err)
		}
	}

	return nil
}

// handleMessage обрабатывает входящие сообщения
func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	// Создаем контекст сообщения
	ctx := &handlers.MessageContext{
		Bot:     b.api,
		Message: message,
		Logger:  b.logger,
	}

	// Обрабатываем команды
	if message.IsCommand() {
		return b.handlers.HandleCommand(ctx)
	}

	// Обрабатываем обычные сообщения
	return b.handlers.HandleMessage(ctx)
}

// Stop останавливает бота
func (b *Bot) Stop() {
	b.logger.Info("Stopping bot...")
}
