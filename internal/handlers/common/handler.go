package common

import (
	"strings"
	"tribute-chatbot/internal/logger"

	tele "gopkg.in/telebot.v4"
)

// Handler обработчик общих команд
type Handler struct {
	logger logger.Logger
}

// NewHandler создает новый обработчик общих команд
func NewHandler() *Handler {
	return &Handler{
		logger: logger.New(),
	}
}

// HandleStart обрабатывает команду /start
func (h *Handler) HandleStart(c tele.Context) error {
	markup := &tele.ReplyMarkup{}
	btn := markup.URL("Get started", "https://t.me/tribute_egorbot/app")
	markup.Inline(markup.Row(btn))
	return c.Send("Welcome! Tribute helps to monetize audiences in Telegram.", markup)
}

// HandleHelp обрабатывает команду /help
func (h *Handler) HandleHelp(c tele.Context) error {
	msg := `📚 Справка по командам:

/start - Начать работу с ботом
/help - Показать эту справку
/echo <текст> - Повторить ваш текст
/verificate - Пройти верификацию (селфи + паспорт)

💡 Просто отправьте мне любое сообщение, и я отвечу!`
	return c.Send(msg)
}

// HandleEcho обрабатывает команду /echo
func (h *Handler) HandleEcho(c tele.Context) error {
	args := c.Message().Payload
	if args == "" {
		return c.Send("Пожалуйста, укажите текст для повторения.\nПример: /echo Привет, мир!")
	}
	return c.Send("🔊 Эхо: " + args)
}

// HandleText обрабатывает обычные текстовые сообщения
func (h *Handler) HandleText(c tele.Context) error {
	text := strings.ToLower(strings.TrimSpace(c.Text()))
	switch {
	case strings.Contains(text, "привет") || strings.Contains(text, "hello"):
		return c.Send("Привет! 👋 Как дела?")
	case strings.Contains(text, "как дела") || strings.Contains(text, "как ты"):
		return c.Send("Спасибо, у меня все отлично! 😊 А у вас как дела?")
	case strings.Contains(text, "спасибо") || strings.Contains(text, "thanks"):
		return c.Send("Пожалуйста! 😊 Рад помочь!")
	case strings.Contains(text, "пока") || strings.Contains(text, "до свидания"):
		return c.Send("До свидания! 👋 Буду ждать нашего следующего разговора!")
	case strings.Contains(text, "время") || strings.Contains(text, "дата"):
		return c.Send("Я не могу показать точное время, но могу помочь с другими вопросами! 🤖")
	default:
		return c.Send("Интересно! Расскажите больше или используйте /help для просмотра команд.")
	}
}

// HandleWebApp обрабатывает WebApp данные
func (h *Handler) HandleWebApp(c tele.Context) error {
	data := c.Message().WebAppData
	if data != nil && data.Data == "verify-account" {
		return c.Send("Account verification data received by bot.")
	}
	return nil
}

// HandleDonate обрабатывает команду /donate
func (h *Handler) HandleDonate(c tele.Context) error {
	photo := &tele.Photo{
		File:    tele.FromDisk("assets/support.jpg"),
		Caption: "<b>Support the Creativity 🌟</b>\nSubscribe to keep our creativity alive! With your help, we can continue creating amazing content just for you. Thank you for being awesome!",
	}
	return c.Send(photo, &tele.SendOptions{ParseMode: tele.ModeHTML})
}
