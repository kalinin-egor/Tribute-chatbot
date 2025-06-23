package main

import (
	"log"
	"strings"
	"sync"
	"time"

	"encoding/json"
	"net/http"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"

	"fmt"

	"strconv"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v4"
)

// VerificationState хранит состояние верификации пользователя
type VerificationState struct {
	UserID     int64
	SelfieID   string
	PassportID string
	Step       string // "waiting_selfie", "waiting_passport", "completed"
}

// VerificationData хранит данные для отправки в админский чат
type VerificationData struct {
	UserID     int64
	SelfieID   string
	PassportID string
	MessageID  int
}

var (
	verificationStates = make(map[int64]*VerificationState)
	verificationMutex  sync.RWMutex
)

// getVerificationState получает состояние верификации пользователя
func getVerificationState(userID int64) *VerificationState {
	verificationMutex.RLock()
	defer verificationMutex.RUnlock()
	return verificationStates[userID]
}

// setVerificationState устанавливает состояние верификации пользователя
func setVerificationState(userID int64, state *VerificationState) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	verificationStates[userID] = state
}

// clearVerificationState очищает состояние верификации пользователя
func clearVerificationState(userID int64) {
	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	delete(verificationStates, userID)
}

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
		btn := markup.URL("Get started", "https://t.me/tribute_egorbot/app")
		markup.Inline(markup.Row(btn))
		return c.Send("Welcome! Tribute helps to monetize audiences in Telegram.", markup)
	})

	// /help
	b.Handle("/help", func(c tele.Context) error {
		msg := `📚 Справка по командам:

/start - Начать работу с ботом
/help - Показать эту справку
/echo <текст> - Повторить ваш текст
/verificate - Пройти верификацию (селфи + паспорт)

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

	// /verificate
	b.Handle("/verificate", func(c tele.Context) error {
		userID := c.Sender().ID

		// Инициализируем состояние верификации
		state := &VerificationState{
			UserID: userID,
			Step:   "waiting_selfie",
		}
		setVerificationState(userID, state)

		return c.Send("🔐 Начинаем процесс верификации!\n\n📸 Пожалуйста, отправьте ваше селфи (фотографию лица).")
	})

	// WebAppData - отдельное событие
	b.Handle(tele.OnWebApp, func(c tele.Context) error {
		data := c.Message().WebAppData
		if data != nil && data.Data == "verify-account" {
			return c.Send("Account verification data received by bot.")
		}
		return nil
	})

	// Обработка фотографий для верификации
	b.Handle(tele.OnPhoto, func(c tele.Context) error {
		userID := c.Sender().ID
		state := getVerificationState(userID)

		if state == nil {
			return c.Send("❌ Сначала используйте команду /verificate для начала процесса верификации.")
		}

		photo := c.Message().Photo
		if photo == nil {
			return c.Send("❌ Не удалось получить фотографию. Попробуйте еще раз.")
		}

		fileID := photo.FileID

		switch state.Step {
		case "waiting_selfie":
			state.SelfieID = fileID
			state.Step = "waiting_passport"
			setVerificationState(userID, state)
			return c.Send("✅ Селфи получено!\n\n📄 Теперь отправьте фотографию паспорта (страница с фото и данными).")

		case "waiting_passport":
			state.PassportID = fileID
			state.Step = "completed"
			setVerificationState(userID, state)

			// Отправляем фотографии в админский чат
			return sendVerificationToAdmin(b, c, state, cfg.TelegramAdminChatID)

		default:
			return c.Send("❌ Неожиданное состояние. Используйте /verificate для начала заново.")
		}
	})

	// Обработка callback кнопок верификации
	b.Handle(tele.OnCallback, func(c tele.Context) error {
		callback := c.Callback()
		if callback == nil {
			logg.Error("Callback is nil")
			return nil
		}

		data := callback.Data
		logg.Info(fmt.Sprintf("Received callback data: '%s' from user: %d", data, callback.Sender.ID))

		if strings.HasPrefix(data, "verify_user_") {
			logg.Info("Processing verification callback")
			return handleVerificationCallback(b, c, data, client, cfg)
		} else {
			logg.Info("Callback data does not match verify_user_ pattern")
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

// sendVerificationToAdmin отправляет фотографии верификации в админский чат
func sendVerificationToAdmin(bot *tele.Bot, c tele.Context, state *VerificationState, adminChatID int64) error {
	logg := logger.New()
	adminChat := &tele.Chat{ID: adminChatID}

	logg.Info(fmt.Sprintf("Sending verification to admin chat: %d", adminChatID))

	// Проверяем, что бот может отправлять сообщения в админский чат
	testMsg := "🔧 Тест отправки сообщений"
	_, err := bot.Send(adminChat, testMsg)
	if err != nil {
		logg.Error("Bot cannot send messages to admin chat:", err)
		return c.Send("❌ Ошибка: бот не может отправлять сообщения в админский чат. Проверьте права бота.")
	}

	// Создаем inline кнопки
	markup := bot.NewMarkup()
	approveData := fmt.Sprintf("verify_user_%d_true", state.UserID)
	rejectData := fmt.Sprintf("verify_user_%d_false", state.UserID)

	logg.Info(fmt.Sprintf("Creating buttons with data: approve='%s', reject='%s'", approveData, rejectData))

	approveBtn := markup.Data("✅ Подтвердить", approveData)
	rejectBtn := markup.Data("❌ Отозвать", rejectData)
	markup.Inline(markup.Row(approveBtn, rejectBtn))

	// Отправляем селфи с кнопками
	selfieMsg := &tele.Photo{
		File:    tele.File{FileID: state.SelfieID},
		Caption: fmt.Sprintf("🔐 Заявка на верификацию\n👤 Пользователь: %d\n📸 Селфи", state.UserID),
	}
	_, err = bot.Send(adminChat, selfieMsg, markup)
	if err != nil {
		logg.Error("Failed to send selfie with buttons:", err)
		return c.Send("❌ Ошибка при отправке заявки. Попробуйте позже.")
	}

	logg.Info("Successfully sent selfie with buttons")

	// Отправляем паспорт
	passportMsg := &tele.Photo{
		File:    tele.File{FileID: state.PassportID},
		Caption: "📄 Фотография паспорта",
	}
	_, err = bot.Send(adminChat, passportMsg)
	if err != nil {
		logg.Error("Failed to send passport:", err)
		return c.Send("❌ Ошибка при отправке заявки. Попробуйте позже.")
	}

	logg.Info("Successfully sent passport")

	// Очищаем состояние
	clearVerificationState(state.UserID)

	return c.Send("✅ Ваша заявка на верификацию отправлена администратору!\n\n⏳ Ожидайте решения. Мы уведомим вас о результате.")
}

// handleVerificationCallback обрабатывает нажатие кнопок верификации
func handleVerificationCallback(bot *tele.Bot, c tele.Context, data string, client *http.Client, cfg *config.Config) error {
	logg := logger.New()

	// Парсим данные: verify_user_<user_id>_<true/false>
	parts := strings.Split(data, "_")
	if len(parts) != 4 {
		logg.Error("Invalid callback data format:", data)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка обработки запроса"})
	}

	userIDStr := parts[2]
	verificationStatus := parts[3]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		logg.Error("Failed to parse user ID:", userIDStr, err)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка обработки запроса"})
	}

	isVerified := verificationStatus == "true"

	logg.Info(fmt.Sprintf("Processing verification callback: user_id=%d, verified=%t", userID, isVerified))

	// Отправляем запрос к API
	payload := map[string]interface{}{
		"userId":        userID,
		"isVerificated": isVerified,
	}

	body, _ := json.Marshal(payload)
	apiURL := strings.TrimRight(cfg.APIBaseURL, "/") + "/v1/check-verified-passport"
	req, _ := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	logg.Info("Sending API request to:", apiURL)

	resp, err := client.Do(req)
	if err != nil {
		logg.Error("API request failed:", err)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка при обновлении статуса верификации"})
	}
	defer resp.Body.Close()

	logg.Info("API response status:", resp.StatusCode)

	// Проверяем статус ответа
	if resp.StatusCode != 200 {
		logg.Error("API returned non-200 status:", resp.StatusCode)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка сервера при обновлении статуса"})
	}

	// Удаляем сообщения с фотографиями из админского чата
	callback := c.Callback()
	if callback != nil && callback.Message != nil {
		// Удаляем текущее сообщение (с кнопками)
		err = bot.Delete(callback.Message)
		if err != nil {
			logg.Error("Failed to delete message with buttons:", err)
		} else {
			logg.Info("Successfully deleted message with buttons")
		}

		// Удаляем предыдущее сообщение (с паспортом)
		if callback.Message.ID > 1 {
			prevMsg := &tele.Message{
				ID:   callback.Message.ID - 1,
				Chat: callback.Message.Chat,
			}
			err = bot.Delete(prevMsg)
			if err != nil {
				logg.Error("Failed to delete passport message:", err)
			} else {
				logg.Info("Successfully deleted passport message")
			}
		}
	}

	// Отправляем уведомление пользователю
	userChat := &tele.Chat{ID: userID}
	statusText := "✅ Верификация подтверждена!"
	if !isVerified {
		statusText = "❌ Верификация отклонена"
	}

	_, err = bot.Send(userChat, statusText)
	if err != nil {
		logg.Error("Failed to send notification to user:", err)
	}

	logg.Info(fmt.Sprintf("Verification processed successfully: user_id=%d, verified=%t", userID, isVerified))

	// Отвечаем на callback
	return c.Respond(&tele.CallbackResponse{Text: "✅ Статус верификации обновлен"})
}
