package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger интерфейс для логирования
type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})
	WithField(key string, value interface{}) Logger
}

// logger реализация логгера
type logger struct {
	log *logrus.Logger
}

// New создает новый экземпляр логгера
func New() Logger {
	log := logrus.New()

	// Настраиваем формат логов
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Устанавливаем уровень логирования
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	return &logger{log: log}
}

func (l *logger) Info(args ...interface{}) {
	l.log.Info(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.log.Error(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.log.Fatal(args...)
}

func (l *logger) Debug(args ...interface{}) {
	l.log.Debug(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.log.Warn(args...)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{log: l.log.WithField(key, value).Logger}
}
