package main

import (
	"log"
	"strings"
	"time"

	"encoding/json"
	"net/http"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	"fmt"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v4"
)

func main() {
	godotenv.Load()

	logg := logger.New()
	cfg, err := config.Load()
	if err != nil {
		logg.Fatal("Failed to load configuration", err)
	}

	pref := tele.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	}
	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// /start
	b.Handle("/start", func(c tele.Context) error {
		markup := b.NewMarkup()
		btn := markup.WebApp("Get started", &tele.WebApp{URL: "https://t.me/tribute_egorbot/app"})
		markup.Inline(markup.Row(btn))
		return c.Send("Welcome! Tribute helps to monetize audiences in Telegram.", markup)
	})

	// /help
	b.Handle("/help", func(c tele.Context) error {
		msg := `📚 Справка по командам:

/start - Начать работу с ботом
/help - Показать эту справку
/echo <текст> - Повторить ваш текст

💡 Просто отправьте мне любое сообщение, и я отвечу!`
		return c.Send(msg)
	})

	// /echo
	b.Handle("/echo", func(c tele.Context) error {
		args := c.Message().Payload
		if args == "" {
			return c.Send("Пожалуйста, укажите текст для повторения.\nПример: /echo Привет, мир!")
		}
		return c.Send("🔊 Эхо: " + args)
	})

	// WebAppData - отдельное событие
	b.Handle(tele.OnWebApp, func(c tele.Context) error {
		data := c.Message().WebAppData
		if data != nil && data.Data == "verify-account" {
			return c.Send("Account verification data received by bot.")
		}
		return nil
	})

	// Обычные сообщения (AI-ответы)
	b.Handle(tele.OnText, func(c tele.Context) error {
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
	})

	// my_chat_member
	b.Handle(tele.OnMyChatMember, func(c tele.Context) error {
		upd := c.ChatMember()
		oldStatus := upd.OldChatMember.Role
		newStatus := upd.NewChatMember.Role

		logg.Info(fmt.Sprintf("my_chat_member: chat_id=%d, user_id=%d, old=%s, new=%s", upd.Chat.ID, upd.NewChatMember.User.ID, oldStatus, newStatus))

		// Если бот стал админом
		if oldStatus != "administrator" && newStatus == "administrator" {
			userID := upd.NewChatMember.User.ID
			channelTitle := upd.Chat.Title
			channelUsername := upd.Chat.Username

			payload := map[string]interface{}{
				"user_id":          userID,
				"channel_title":    channelTitle,
				"channel_username": channelUsername,
			}
			body, _ := json.Marshal(payload)
			apiURL := strings.TrimRight(cfg.APIBaseURL, "/") + "/v1/add-bot"
			req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				logg.Error("API request failed:", err)
				return nil
			}
			defer resp.Body.Close()
			if resp.StatusCode == 400 {
				b.Send(upd.Sender, "Channel is already added")
			}
			return nil
		}

		logg.Info("my_chat_member update: ", oldStatus, " -> ", newStatus)
		return nil
	})

	logg.Info("Starting Telegram bot (Telebot)...")
	b.Start()
}
