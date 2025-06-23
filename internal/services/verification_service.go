package services

import (
	"sync"
	"tribute-chatbot/internal/models"
)

// VerificationService сервис для управления верификацией
type VerificationService struct {
	states map[int64]*models.VerificationState
	mutex  sync.RWMutex
}

// NewVerificationService создает новый сервис верификации
func NewVerificationService() *VerificationService {
	return &VerificationService{
		states: make(map[int64]*models.VerificationState),
	}
}

// GetState получает состояние верификации пользователя
func (s *VerificationService) GetState(userID int64) *models.VerificationState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.states[userID]
}

// SetState устанавливает состояние верификации пользователя
func (s *VerificationService) SetState(userID int64, state *models.VerificationState) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.states[userID] = state
}

// ClearState очищает состояние верификации пользователя
func (s *VerificationService) ClearState(userID int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.states, userID)
}

// InitializeState инициализирует состояние верификации для пользователя
func (s *VerificationService) InitializeState(userID int64) *models.VerificationState {
	state := &models.VerificationState{
		UserID: userID,
		Step:   models.VerificationStepWaitingSelfie,
	}
	s.SetState(userID, state)
	return state
}

// UpdateSelfie обновляет селфи в состоянии верификации
func (s *VerificationService) UpdateSelfie(userID int64, selfieID string) *models.VerificationState {
	state := s.GetState(userID)
	if state != nil {
		state.SelfieID = selfieID
		state.Step = models.VerificationStepWaitingPassport
		s.SetState(userID, state)
	}
	return state
}

// UpdatePassport обновляет паспорт в состоянии верификации
func (s *VerificationService) UpdatePassport(userID int64, passportID string) *models.VerificationState {
	state := s.GetState(userID)
	if state != nil {
		state.PassportID = passportID
		state.Step = models.VerificationStepCompleted
		s.SetState(userID, state)
	}
	return state
}

// UpdateMessageIDs обновляет ID сообщений в состоянии верификации
func (s *VerificationService) UpdateMessageIDs(userID int64, selfieMessageID, passportMessageID int) {
	state := s.GetState(userID)
	if state != nil {
		state.SelfieMessageID = selfieMessageID
		state.PassportMessageID = passportMessageID
		s.SetState(userID, state)
	}
}
