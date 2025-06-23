package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageContext содержит контекст для обработки сообщения
type MessageContext struct {
	Bot     *tgbotapi.BotAPI
	Message *tgbotapi.Message
	Logger  logger.Logger
}

// ChatMemberContext содержит контекст для обработки изменения статуса участника
type ChatMemberContext struct {
	Bot    *tgbotapi.BotAPI
	Update *tgbotapi.ChatMemberUpdated
	Logger logger.Logger
}

// Handlers содержит все обработчики сообщений
type Handlers struct {
	logger logger.Logger
	config *config.Config
	client *http.Client
}

// New создает новый экземпляр обработчиков
func New(cfg *config.Config, log logger.Logger) *Handlers {
	return &Handlers{
		logger: log,
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// HandleCommand обрабатывает команды
func (h *Handlers) HandleCommand(ctx *MessageContext) error {
	command := ctx.Message.Command()
	args := ctx.Message.CommandArguments()

	h.logger.Debug("Handling command:", command, "with args:", args)

	switch command {
	case "start":
		return h.handleStart(ctx)
	case "help":
		return h.handleHelp(ctx)
	case "echo":
		return h.handleEcho(ctx, args)
	default:
		return h.handleUnknownCommand(ctx)
	}
}

// HandleMessage обрабатывает обычные сообщения
func (h *Handlers) HandleMessage(ctx *MessageContext) error {
	text := ctx.Message.Text
	h.logger.Debug("Handling message:", text)

	// Простая логика обработки сообщений
	response := h.processMessage(text)
	return h.sendResponse(ctx, response)
}

// HandleMyChatMember обрабатывает изменение статуса бота в чате
func (h *Handlers) HandleMyChatMember(ctx *ChatMemberContext) error {
	chat := ctx.Update.Chat
	oldStatus := ctx.Update.OldChatMember.Status
	newStatus := ctx.Update.NewChatMember.Status

	log := h.logger.WithField("chat_id", chat.ID).WithField("chat_title", chat.Title)

	log.Info("Bot status changed from '", oldStatus, "' to '", newStatus, "' in chat")

	// Проверяем, стал ли бот администратором
	wasAddedAsAdmin := (oldStatus == "left" || oldStatus == "kicked") && newStatus == "administrator"
	wasPromoted := oldStatus == "member" && newStatus == "administrator"

	if wasAddedAsAdmin || wasPromoted {
		log.Info("Bot is now an administrator. Notifying API...")
		if err := h.notifyAPI(ctx); err != nil {
			log.Error("Failed to notify API:", err)
			// Не возвращаем ошибку дальше, чтобы не прерывать работу бота из-за API
		}
	} else if (oldStatus == "left" || oldStatus == "kicked") && newStatus == "member" {
		log.Info("Bot was added as a member.")
	} else if newStatus == "left" || newStatus == "kicked" {
		log.Warn("Bot was removed from the chat.")
	} else if oldStatus == "administrator" && newStatus == "member" {
		log.Warn("Bot was demoted from administrator.")
	}

	return nil
}

// notifyAPI отправляет уведомление на внешний API
func (h *Handlers) notifyAPI(ctx *ChatMemberContext) error {
	log := h.logger.WithField("chat_id", ctx.Update.Chat.ID)

	channelUsername := ""
	if ctx.Update.Chat.UserName != "" {
		channelUsername = "@" + ctx.Update.Chat.UserName
	}

	// Формируем payload
	payload := struct {
		UserID          int64  `json:"user_id"`
		ChannelTitle    string `json:"channel_title"`
		ChannelUsername string `json:"channel_username"`
	}{
		UserID:          ctx.Update.From.ID,
		ChannelTitle:    ctx.Update.Chat.Title,
		ChannelUsername: channelUsername,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Формируем URL
	url := fmt.Sprintf("%s/v1/add-bot", h.config.APIBaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create api request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	log.Info("Sending notification to API: ", url)

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("api returned status %s", resp.Status)
	}

	log.Info("Successfully notified API. Status: ", resp.Status)

	return nil
}

// handleStart обрабатывает команду /start
func (h *Handlers) handleStart(ctx *MessageContext) error {
	message := "Welcome! Tribute helps to monetize audiences in Telegram."

	// Создаем сообщение с текстом
	msg := tgbotapi.NewMessage(ctx.Message.Chat.ID, message)
	msg.ParseMode = "HTML"

	// Создаем inline клавиатуру с URL кнопкой
	url := "https://t.me/tribute_egorbot/app"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("Get started", url),
		),
	)
	msg.ReplyMarkup = keyboard

	_, err := ctx.Bot.Send(msg)
	if err != nil {
		h.logger.Error("Failed to send start message with inline button:", err)
		return fmt.Errorf("failed to send start message with inline button: %w", err)
	}

	return nil
}

// handleHelp обрабатывает команду /help
func (h *Handlers) handleHelp(ctx *MessageContext) error {
	message := `📚 Справка по командам:

/start - Начать работу с ботом
/help - Показать эту справку
/echo <текст> - Повторить ваш текст

💡 Просто отправьте мне любое сообщение, и я отвечу!`

	return h.sendResponse(ctx, message)
}

// handleEcho обрабатывает команду /echo
func (h *Handlers) handleEcho(ctx *MessageContext, args string) error {
	if args == "" {
		return h.sendResponse(ctx, "Пожалуйста, укажите текст для повторения.\nПример: /echo Привет, мир!")
	}

	return h.sendResponse(ctx, fmt.Sprintf("🔊 Эхо: %s", args))
}

// handleUnknownCommand обрабатывает неизвестные команды
func (h *Handlers) handleUnknownCommand(ctx *MessageContext) error {
	message := fmt.Sprintf(
		"❓ Неизвестная команда: /%s\n\nИспользуйте /help для просмотра доступных команд.",
		ctx.Message.Command(),
	)

	return h.sendResponse(ctx, message)
}

// processMessage обрабатывает обычные сообщения
func (h *Handlers) processMessage(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))

	switch {
	case strings.Contains(text, "привет") || strings.Contains(text, "hello"):
		return "Привет! 👋 Как дела?"
	case strings.Contains(text, "как дела") || strings.Contains(text, "как ты"):
		return "Спасибо, у меня все отлично! 😊 А у вас как дела?"
	case strings.Contains(text, "спасибо") || strings.Contains(text, "thanks"):
		return "Пожалуйста! 😊 Рад помочь!"
	case strings.Contains(text, "пока") || strings.Contains(text, "до свидания"):
		return "До свидания! 👋 Буду ждать нашего следующего разговора!"
	case strings.Contains(text, "время") || strings.Contains(text, "дата"):
		return "Я не могу показать точное время, но могу помочь с другими вопросами! 🤖"
	default:
		return "Интересно! Расскажите больше или используйте /help для просмотра команд."
	}
}

// sendResponse отправляет ответ пользователю
func (h *Handlers) sendResponse(ctx *MessageContext, text string) error {
	msg := tgbotapi.NewMessage(ctx.Message.Chat.ID, text)
	msg.ParseMode = "HTML"

	_, err := ctx.Bot.Send(msg)
	if err != nil {
		h.logger.Error("Failed to send message:", err)
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
