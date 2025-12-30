package jobs

import (
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/service"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/telegram"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type JobScheduler struct {
	db            *gorm.DB
	accountingSvc service.AccountingService
	userRepo      repository.UserRepository
	usageRepo     repository.UsageRepository
	thresholdRepo *thresholdRepository
	telegramBot   *telegram.Bot
	logger        *zap.Logger
}

type thresholdRepository struct {
	db *gorm.DB
}

func NewJobScheduler(
	db *gorm.DB,
	accountingSvc service.AccountingService,
	userRepo repository.UserRepository,
	usageRepo repository.UsageRepository,
	telegramBot *telegram.Bot,
	logger *zap.Logger,
) *JobScheduler {
	return &JobScheduler{
		db:            db,
		accountingSvc: accountingSvc,
		userRepo:      userRepo,
		usageRepo:     usageRepo,
		thresholdRepo: &thresholdRepository{db: db},
		telegramBot:   telegramBot,
		logger:        logger,
	}
}

func (s *JobScheduler) Start() {
	// Plan reset job - runs every hour
	go s.runPeriodic("plan_reset", 1*time.Hour, s.checkPlanResets)

	// Telegram notifications - runs every 5 minutes
	go s.runPeriodic("telegram_notifications", 5*time.Minute, s.checkNotificationThresholds)

	// Online users cleanup - runs every 10 minutes
	go s.runPeriodic("online_cleanup", 10*time.Minute, s.cleanupStaleOnlineUsers)

	s.logger.Info("Background jobs started")
}

func (s *JobScheduler) runPeriodic(name string, interval time.Duration, job func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.logger.Info("Running background job", zap.String("job", name))
		job()
	}
}

func (s *JobScheduler) checkPlanResets() {
	// This would check all current periods and reset if needed
	// For now, it's a placeholder
	s.logger.Debug("Checking plan resets")

	// TODO: Implement actual plan reset logic
	// 1. Query all current periods where period_end < now
	// 2. Close those periods (set is_current = false)
	// 3. Create new periods for those users
}

func (s *JobScheduler) checkNotificationThresholds() {
	if s.telegramBot == nil {
		return
	}

	s.logger.Debug("Checking notification thresholds")

	// Get all users with Telegram linked and active thresholds
	// This is simplified - in production you'd query more efficiently
	users, _, err := s.userRepo.List(0, 10000)
	if err != nil {
		s.logger.Error("Failed to list users for notifications", zap.Error(err))
		return
	}

	for _, user := range users {
		if user.TelegramChatID == nil {
			continue
		}

		// Get current usage
		usage, err := s.usageRepo.GetCurrentPeriod(user.ID)
		if err != nil {
			continue
		}

		if user.Plan == nil {
			continue
		}

		totalBillable := usage.BillableBytesUp + usage.BillableBytesDown
		percentUsed := float64(totalBillable) / float64(user.Plan.QuotaBytes) * 100

		// Check if any threshold is crossed
		// TODO: Load thresholds from database and check against them
		// For now, we'll use hardcoded thresholds
		thresholds := []float64{50, 80, 95}
		for _, threshold := range thresholds {
			if percentUsed >= threshold && percentUsed < threshold+5 {
				// Send notification
				message := telegram.FormatUsageNotification(
					user.Email,
					usage.RealBytesUp,
					usage.RealBytesDown,
					usage.BillableBytesUp,
					usage.BillableBytesDown,
					user.Plan.QuotaBytes,
					percentUsed,
				)

				if err := s.telegramBot.SendNotification(*user.TelegramChatID, message, "threshold"); err != nil {
					s.logger.Error("Failed to send threshold notification",
						zap.Uint64("user_id", user.ID),
						zap.Error(err),
					)
				}
			}
		}
	}
}

func (s *JobScheduler) cleanupStaleOnlineUsers() {
	s.logger.Debug("Cleaning up stale online users")

	// Delete online users not seen in the last 5 minutes
	// This would be implemented in the repository
	// For now, it's a placeholder
}
