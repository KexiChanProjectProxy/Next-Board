package telegram

import (
	"fmt"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/config"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/metrics"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	bot      *tgbotapi.BotAPI
	userRepo repository.UserRepository
	logger   *zap.Logger
}

func NewBot(cfg *config.TelegramConfig, userRepo repository.UserRepository, logger *zap.Logger) (*Bot, error) {
	if cfg.Token == "" {
		logger.Warn("Telegram token not configured, bot will not start")
		return nil, nil
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, err
	}

	logger.Info("Telegram bot authorized", zap.String("username", bot.Self.UserName))

	return &Bot{
		bot:      bot,
		userRepo: userRepo,
		logger:   logger,
	}, nil
}

func (b *Bot) Start() {
	if b == nil || b.bot == nil {
		return
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.sendMessage(message.Chat.ID, "Welcome! Use /link <token> to link your account.")
	case "link":
		// TODO: Implement token-based linking
		args := message.CommandArguments()
		if args == "" {
			b.sendMessage(message.Chat.ID, "Usage: /link <token>")
			return
		}
		b.sendMessage(message.Chat.ID, "Token linking not yet implemented. Token: "+args)
	case "status":
		// Check if user is linked
		user, err := b.userRepo.FindByTelegramChatID(message.Chat.ID)
		if err != nil {
			b.sendMessage(message.Chat.ID, "Your account is not linked. Use /link <token> to link.")
			return
		}
		b.sendMessage(message.Chat.ID, fmt.Sprintf("Your account (%s) is linked!", user.Email))
	default:
		b.sendMessage(message.Chat.ID, "Unknown command. Available commands: /start, /link, /status")
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.Error("Failed to send Telegram message", zap.Error(err))
	}
}

func (b *Bot) SendNotification(chatID int64, message string, notificationType string) error {
	if b == nil || b.bot == nil {
		return fmt.Errorf("bot not initialized")
	}

	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.bot.Send(msg)

	if err == nil {
		metrics.TelegramNotificationsTotal.WithLabelValues(notificationType).Inc()
	}

	return err
}

func FormatUsageNotification(email string, realUp, realDown, billableUp, billableDown, quota uint64, percentUsed float64) string {
	return fmt.Sprintf(
		"Usage Alert for %s\n\n"+
			"Real Usage:\n"+
			"  Upload: %s\n"+
			"  Download: %s\n"+
			"  Total: %s\n\n"+
			"Billable Usage:\n"+
			"  Upload: %s\n"+
			"  Download: %s\n"+
			"  Total: %s\n\n"+
			"Quota: %s\n"+
			"Used: %.1f%%",
		email,
		formatBytes(realUp),
		formatBytes(realDown),
		formatBytes(realUp+realDown),
		formatBytes(billableUp),
		formatBytes(billableDown),
		formatBytes(billableUp+billableDown),
		formatBytes(quota),
		percentUsed,
	)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
