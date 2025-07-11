package channel

import (
	"fmt"
	"strings"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"
	"tribute-chatbot/internal/services"

	tele "gopkg.in/telebot.v4"
)

// Handler обработчик каналов
type Handler struct {
	apiService *services.APIService
	config     *config.Config
	logger     logger.Logger
}

// NewHandler создает новый обработчик каналов
func NewHandler(apiService *services.APIService, config *config.Config) *Handler {
	return &Handler{
		apiService: apiService,
		config:     config,
		logger:     logger.New(),
	}
}

// HandleMyChatMember обрабатывает события добавления бота в каналы
func (h *Handler) HandleMyChatMember(c tele.Context) error {
	upd := c.ChatMember()
	oldStatus := strings.ToLower(strings.TrimSpace(string(upd.OldChatMember.Role)))
	newStatus := strings.ToLower(strings.TrimSpace(string(upd.NewChatMember.Role)))

	h.logger.Info(fmt.Sprintf("my_chat_member: chat_id=%d, user_id=%d, old=%s, new=%s",
		upd.Chat.ID, upd.NewChatMember.User.ID, oldStatus, newStatus))

	// Если бот стал админом
	if oldStatus != "administrator" && newStatus == "administrator" {
		userID := c.Sender().ID
		channelTitle := upd.Chat.Title
		channelUsername := upd.Chat.Username

		h.logger.Info(fmt.Sprintf("userID: %d, channelTitle: '%s', channelUsername: '%s'", userID, channelTitle, channelUsername))

		if userID == 0 || channelTitle == "" {
			h.logger.Error("userID or channelTitle is empty, not calling AddBotToChannel")
			return nil
		}

		h.logger.Info("Calling AddBotToChannel endpoint...")
		err := h.apiService.AddBotToChannel(userID, channelTitle, channelUsername)
		if err != nil {
			if strings.Contains(err.Error(), "channel is already added") {
				c.Bot().Send(upd.Sender, "Channel is already added")
			} else {
				h.logger.Error("Failed to add bot to channel:", err)
			}
		}
		return nil
	}

	h.logger.Info("my_chat_member update: ", oldStatus, " -> ", newStatus)
	return nil
}
