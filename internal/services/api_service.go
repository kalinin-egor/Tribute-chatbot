package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/logger"
)

// APIService сервис для работы с API
type APIService struct {
	client *http.Client
	config *config.Config
	logger logger.Logger
}

// NewAPIService создает новый API сервис
func NewAPIService(cfg *config.Config) *APIService {
	return &APIService{
		client: &http.Client{Timeout: 10 * time.Second},
		config: cfg,
		logger: logger.New(),
	}
}

// UpdateUserVerification обновляет статус верификации пользователя
func (s *APIService) UpdateUserVerification(userID int64, isVerified bool) error {
	payload := map[string]interface{}{
		"userId":        userID,
		"isVerificated": isVerified,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	apiURL := strings.TrimRight(s.config.APIBaseURL, "/") + "/v1/check-verified-passport"
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	s.logger.Info("Sending API request to:", apiURL)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Info("API response status:", resp.StatusCode)

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// AddBotToChannel добавляет бота в канал
func (s *APIService) AddBotToChannel(userID int64, channelTitle, channelUsername string) error {
	s.logger.Info(fmt.Sprintf("AddBotToChannel called with: userID=%d, channelTitle='%s', channelUsername='%s'", userID, channelTitle, channelUsername))

	payload := map[string]interface{}{
		"user_id":          userID,
		"channel_title":    channelTitle,
		"channel_username": channelUsername,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	apiURL := strings.TrimRight(s.config.APIBaseURL, "/") + "/v1/add-bot"
	s.logger.Info("AddBotToChannel API URL:", apiURL)
	s.logger.Info("AddBotToChannel payload:", string(body))

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	s.logger.Info("AddBotToChannel response status:", resp.StatusCode)

	if resp.StatusCode == 400 {
		return fmt.Errorf("channel is already added")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}
