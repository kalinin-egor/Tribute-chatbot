package bot

import (
	"time"
	"tribute-chatbot/internal/config"
	"tribute-chatbot/internal/handlers/channel"
	"tribute-chatbot/internal/handlers/common"
	"tribute-chatbot/internal/handlers/verification"
	"tribute-chatbot/internal/logger"
	"tribute-chatbot/internal/services"

	tele "gopkg.in/telebot.v4"
)

// Bot основная структура бота
type Bot struct {
	bot                 *tele.Bot
	config              *config.Config
	logger              logger.Logger
	verificationService *services.VerificationService
	apiService          *services.APIService
	commonHandler       *common.Handler
	verificationHandler *verification.Handler
	channelHandler      *channel.Handler
}

// NewBot создает новый экземпляр бота
func NewBot(cfg *config.Config) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 30 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	// Инициализируем сервисы
	verificationService := services.NewVerificationService()
	apiService := services.NewAPIService(cfg)

	// Инициализируем обработчики
	commonHandler := common.NewHandler()
	verificationHandler := verification.NewHandler(verificationService, apiService, cfg)
	channelHandler := channel.NewHandler(apiService, cfg)

	return &Bot{
		bot:                 bot,
		config:              cfg,
		logger:              logger.New(),
		verificationService: verificationService,
		apiService:          apiService,
		commonHandler:       commonHandler,
		verificationHandler: verificationHandler,
		channelHandler:      channelHandler,
	}, nil
}

// SetupHandlers настраивает обработчики команд и событий
func (b *Bot) SetupHandlers() {
	// Общие команды
	b.bot.Handle("/start", b.commonHandler.HandleStart)
	b.bot.Handle("/help", b.commonHandler.HandleHelp)
	b.bot.Handle("/echo", b.commonHandler.HandleEcho)
	b.bot.Handle("/donate", b.commonHandler.HandleDonate)

	// Верификация
	b.bot.Handle("/verificate", b.verificationHandler.HandleStartVerification)
	b.bot.Handle(tele.OnPhoto, b.verificationHandler.HandlePhoto)
	b.bot.Handle(tele.OnCallback, b.verificationHandler.HandleCallback)

	// WebApp
	b.bot.Handle(tele.OnWebApp, b.commonHandler.HandleWebApp)

	// Текстовые сообщения
	b.bot.Handle(tele.OnText, b.commonHandler.HandleText)

	// Каналы
	b.bot.Handle(tele.OnMyChatMember, b.channelHandler.HandleMyChatMember)
}

// Start запускает бота
func (b *Bot) Start() {
	b.logger.Info("Starting Telegram bot (Telebot)...")
	b.SetupHandlers()
	b.bot.Start()
}

// Stop останавливает бота
func (b *Bot) Stop() {
	b.bot.Stop()
}
