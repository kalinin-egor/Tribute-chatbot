package handlers

import (
	"fmt"
	"strings"

	"tribute-chatbot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// MessageContext содержит контекст для обработки сообщения
type MessageContext struct {
	Bot     *tgbotapi.BotAPI
	Message *tgbotapi.Message
	Logger  logger.Logger
}

// Handlers содержит все обработчики сообщений
type Handlers struct {
	logger logger.Logger
}

// New создает новый экземпляр обработчиков
func New(log logger.Logger) *Handlers {
	return &Handlers{
		logger: log,
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

// handleStart обрабатывает команду /start
func (h *Handlers) handleStart(ctx *MessageContext) error {
	message := fmt.Sprintf(
		"Привет, %s! 👋\n\nЯ Tribute Chatbot - ваш помощник.\n\n"+
			"Доступные команды:\n"+
			"/start - показать это сообщение\n"+
			"/help - показать справку\n"+
			"/echo <текст> - повторить текст\n\n"+
			"Просто напишите мне сообщение, и я отвечу!",
		ctx.Message.From.FirstName,
	)

	return h.sendResponse(ctx, message)
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
