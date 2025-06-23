package models

// VerificationState хранит состояние верификации пользователя
type VerificationState struct {
	UserID            int64
	SelfieID          string
	PassportID        string
	Step              string // "waiting_selfie", "waiting_passport", "completed"
	SelfieMessageID   int
	PassportMessageID int
}

// VerificationData хранит данные для отправки в админский чат
type VerificationData struct {
	UserID     int64
	SelfieID   string
	PassportID string
	MessageID  int
}

// VerificationCallbackData данные callback кнопки верификации
type VerificationCallbackData struct {
	UserID        int64
	IsVerificated bool
}

// VerificationStep этапы верификации
const (
	VerificationStepWaitingSelfie   = "waiting_selfie"
	VerificationStepWaitingPassport = "waiting_passport"
	VerificationStepCompleted       = "completed"
)
