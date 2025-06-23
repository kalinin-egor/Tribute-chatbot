package verification

import (
	"fmt"
	"strconv"
	"strings"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"
	"tribute-chatbot/internal/models"
	"tribute-chatbot/internal/services"

	tele "gopkg.in/telebot.v4"
)

// Handler обработчик верификации
type Handler struct {
	verificationService *services.VerificationService
	apiService          *services.APIService
	config              *config.Config
	logger              logger.Logger
}

// NewHandler создает новый обработчик верификации
func NewHandler(
	verificationService *services.VerificationService,
	apiService *services.APIService,
	config *config.Config,
) *Handler {
	return &Handler{
		verificationService: verificationService,
		apiService:          apiService,
		config:              config,
		logger:              logger.New(),
	}
}

// HandleStartVerification обрабатывает команду /verificate
func (h *Handler) HandleStartVerification(c tele.Context) error {
	userID := c.Sender().ID

	// Инициализируем состояние верификации
	h.verificationService.InitializeState(userID)

	return c.Send("🔐 Начинаем процесс верификации!\n\n📸 Пожалуйста, отправьте ваше селфи (фотографию лица).")
}

// HandlePhoto обрабатывает фотографии для верификации
func (h *Handler) HandlePhoto(c tele.Context) error {
	userID := c.Sender().ID
	state := h.verificationService.GetState(userID)

	if state == nil {
		return c.Send("❌ Сначала используйте команду /verificate для начала процесса верификации.")
	}

	photo := c.Message().Photo
	if photo == nil {
		return c.Send("❌ Не удалось получить фотографию. Попробуйте еще раз.")
	}

	fileID := photo.FileID

	switch state.Step {
	case models.VerificationStepWaitingSelfie:
		h.verificationService.UpdateSelfie(userID, fileID)
		return c.Send("✅ Селфи получено!\n\n📄 Теперь отправьте фотографию паспорта (страница с фото и данными).")

	case models.VerificationStepWaitingPassport:
		h.verificationService.UpdatePassport(userID, fileID)
		return h.sendVerificationToAdmin(c, state)

	default:
		return c.Send("❌ Неожиданное состояние. Используйте /verificate для начала заново.")
	}
}

// HandleCallback обрабатывает callback кнопки верификации
func (h *Handler) HandleCallback(c tele.Context) error {
	callback := c.Callback()
	if callback == nil {
		h.logger.Error("Callback is nil")
		return nil
	}

	data := strings.TrimSpace(callback.Data)
	h.logger.Info(fmt.Sprintf("Received callback data: '%s' from user: %d", data, callback.Sender.ID))

	if strings.HasPrefix(data, "verify_user_") {
		h.logger.Info("Processing verification callback")
		return h.handleVerificationCallback(c, data)
	} else {
		h.logger.Info("Callback data does not match verify_user_ pattern")
	}

	return nil
}

// sendVerificationToAdmin отправляет фотографии верификации в админский чат
func (h *Handler) sendVerificationToAdmin(c tele.Context, state *models.VerificationState) error {
	adminChat := &tele.Chat{ID: h.config.TelegramAdminChatID}

	h.logger.Info(fmt.Sprintf("Sending verification to admin chat: %d", h.config.TelegramAdminChatID))

	// Создаем inline кнопки
	markup := &tele.ReplyMarkup{}
	approveData := fmt.Sprintf("verify_user_%d_true", state.UserID)
	rejectData := fmt.Sprintf("verify_user_%d_false", state.UserID)

	h.logger.Info(fmt.Sprintf("Creating buttons with data: approve='%s', reject='%s'", approveData, rejectData))

	approveBtn := markup.Data("✅ Подтвердить", approveData)
	rejectBtn := markup.Data("❌ Отозвать", rejectData)
	markup.Inline(markup.Row(approveBtn, rejectBtn))

	// Отправляем селфи с кнопками
	selfieMsg := &tele.Photo{
		File:    tele.File{FileID: state.SelfieID},
		Caption: fmt.Sprintf("🔐 Заявка на верификацию\n👤 Пользователь: %d\n📸 Селфи", state.UserID),
	}
	selfieSentMsg, err := c.Bot().Send(adminChat, selfieMsg, markup)
	if err != nil {
		h.logger.Error("Failed to send selfie with buttons:", err)
		return c.Send("❌ Ошибка при отправке заявки. Попробуйте позже.")
	}

	h.logger.Info(fmt.Sprintf("Successfully sent selfie with buttons. Message ID: %d", selfieSentMsg.ID))

	// Отправляем паспорт
	passportMsg := &tele.Photo{
		File:    tele.File{FileID: state.PassportID},
		Caption: "📄 Фотография паспорта",
	}
	passportSentMsg, err := c.Bot().Send(adminChat, passportMsg)
	if err != nil {
		h.logger.Error("Failed to send passport:", err)
		return c.Send("❌ Ошибка при отправке заявки. Попробуйте позже.")
	}

	h.logger.Info(fmt.Sprintf("Successfully sent passport. Message ID: %d", passportSentMsg.ID))

	// Сохраняем ID сообщений для последующего удаления
	h.verificationService.UpdateMessageIDs(state.UserID, selfieSentMsg.ID, passportSentMsg.ID)

	return c.Send("✅ Ваша заявка на верификацию отправлена администратору!\n\n⏳ Ожидайте решения. Мы уведомим вас о результате.")
}

// handleVerificationCallback обрабатывает нажатие кнопок верификации
func (h *Handler) handleVerificationCallback(c tele.Context, data string) error {
	// Парсим данные: verify_user_<user_id>_<true/false>
	parts := strings.Split(data, "_")
	if len(parts) != 4 {
		h.logger.Error("Invalid callback data format:", data)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка обработки запроса"})
	}

	userIDStr := parts[2]
	verificationStatus := parts[3]

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse user ID:", userIDStr, err)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка обработки запроса"})
	}

	isVerified := verificationStatus == "true"

	h.logger.Info(fmt.Sprintf("Processing verification callback: user_id=%d, verified=%t", userID, isVerified))

	// Отправляем запрос к API
	err = h.apiService.UpdateUserVerification(userID, isVerified)
	if err != nil {
		h.logger.Error("Failed to update user verification:", err)
		return c.Respond(&tele.CallbackResponse{Text: "❌ Ошибка при обновлении статуса верификации"})
	}

	// Удаляем сообщения с фотографиями из админского чата
	callback := c.Callback()
	if callback != nil && callback.Message != nil {
		h.logger.Info(fmt.Sprintf("Attempting to delete messages. Current message ID: %d", callback.Message.ID))

		// Получаем состояние верификации для доступа к ID сообщений
		state := h.verificationService.GetState(userID)
		if state != nil && state.SelfieMessageID > 0 && state.PassportMessageID > 0 {
			// Удаляем сообщение с селфи (с кнопками)
			selfieMsg := &tele.Message{
				ID:   state.SelfieMessageID,
				Chat: callback.Message.Chat,
			}
			err = c.Bot().Delete(selfieMsg)
			if err != nil {
				h.logger.Error("Failed to delete selfie message:", err)
			} else {
				h.logger.Info(fmt.Sprintf("Successfully deleted selfie message ID: %d", state.SelfieMessageID))
			}

			// Удаляем сообщение с паспортом
			passportMsg := &tele.Message{
				ID:   state.PassportMessageID,
				Chat: callback.Message.Chat,
			}
			err = c.Bot().Delete(passportMsg)
			if err != nil {
				h.logger.Error("Failed to delete passport message:", err)
			} else {
				h.logger.Info(fmt.Sprintf("Successfully deleted passport message ID: %d", state.PassportMessageID))
			}

			// Очищаем состояние
			h.verificationService.ClearState(userID)
		} else {
			// Fallback: удаляем текущее сообщение и предыдущее
			err = c.Bot().Delete(callback.Message)
			if err != nil {
				h.logger.Error("Failed to delete message with buttons:", err)
			} else {
				h.logger.Info("Successfully deleted message with buttons")
			}

			if callback.Message.ID > 1 {
				prevMsg := &tele.Message{
					ID:   callback.Message.ID - 1,
					Chat: callback.Message.Chat,
				}
				err = c.Bot().Delete(prevMsg)
				if err != nil {
					h.logger.Error("Failed to delete passport message:", err)
				} else {
					h.logger.Info("Successfully deleted passport message")
				}
			}
		}
	} else {
		h.logger.Error("Callback or callback.Message is nil, cannot delete messages")
	}

	// Отправляем уведомление пользователю
	userChat := &tele.Chat{ID: userID}
	statusText := "✅ Верификация подтверждена!"
	if !isVerified {
		statusText = "❌ Верификация отклонена"
	}

	_, err = c.Bot().Send(userChat, statusText)
	if err != nil {
		h.logger.Error("Failed to send notification to user:", err)
	}

	h.logger.Info(fmt.Sprintf("Verification processed successfully: user_id=%d, verified=%t", userID, isVerified))

	// Отвечаем на callback
	return c.Respond(&tele.CallbackResponse{Text: "✅ Статус верификации обновлен"})
}
