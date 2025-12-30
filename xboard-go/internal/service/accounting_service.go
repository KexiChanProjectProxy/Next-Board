package service

import (
	"errors"
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AccountingService interface {
	ProcessTrafficReport(nodeID uint64, reports []models.TrafficReport) error
	CalculateMultiplier(userID, nodeID uint64) (float64, error)
	GetCurrentUsage(userID uint64) (*models.UsagePeriod, error)
	CheckAndResetPeriods() error
	InitializeUserPeriod(userID uint64) error
}

type accountingService struct {
	userRepo   repository.UserRepository
	nodeRepo   repository.NodeRepository
	planRepo   repository.PlanRepository
	usageRepo  repository.UsageRepository
	uuidRepo   repository.UUIDRepository
	logger     *zap.Logger
}

func NewAccountingService(
	userRepo repository.UserRepository,
	nodeRepo repository.NodeRepository,
	planRepo repository.PlanRepository,
	usageRepo repository.UsageRepository,
	uuidRepo repository.UUIDRepository,
	logger *zap.Logger,
) AccountingService {
	return &accountingService{
		userRepo:  userRepo,
		nodeRepo:  nodeRepo,
		planRepo:  planRepo,
		usageRepo: usageRepo,
		logger:    logger,
	}
}

func (s *accountingService) ProcessTrafficReport(nodeID uint64, reports []models.TrafficReport) error {
	node, err := s.nodeRepo.FindByIDWithLabels(nodeID)
	if err != nil {
		return err
	}

	for _, report := range reports {
		if err := s.processUserTraffic(node, report); err != nil {
			s.logger.Error("Failed to process user traffic",
				zap.Uint64("user_id", report.UserID),
				zap.Uint64("node_id", nodeID),
				zap.Error(err),
			)
			continue
		}
	}

	return nil
}

func (s *accountingService) processUserTraffic(node *models.Node, report models.TrafficReport) error {
	user, err := s.userRepo.FindByID(report.UserID)
	if err != nil {
		return err
	}

	if user.Banned {
		return errors.New("user is banned")
	}

	if user.PlanID == nil {
		return errors.New("user has no plan")
	}

	// Ensure user has a current period
	_, err = s.usageRepo.GetCurrentPeriod(user.ID)
	if err == gorm.ErrRecordNotFound {
		if err := s.InitializeUserPeriod(user.ID); err != nil {
			return err
		}
		_, err = s.usageRepo.GetCurrentPeriod(user.ID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Calculate multiplier
	multiplier, err := s.CalculateMultiplier(user.ID, node.ID)
	if err != nil {
		return err
	}

	// Calculate billable traffic
	billableUp := uint64(float64(report.Upload) * multiplier)
	billableDown := uint64(float64(report.Download) * multiplier)

	// Increment usage
	return s.usageRepo.IncrementUsage(
		user.ID,
		node.ID,
		report.Upload,
		report.Download,
		billableUp,
		billableDown,
	)
}

func (s *accountingService) CalculateMultiplier(userID, nodeID uint64) (float64, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return 0, err
	}

	if user.PlanID == nil {
		return 1.0, nil
	}

	node, err := s.nodeRepo.FindByIDWithLabels(nodeID)
	if err != nil {
		return 0, err
	}

	plan, err := s.planRepo.FindByIDWithLabels(*user.PlanID)
	if err != nil {
		return 0, err
	}

	// Start with node multiplier
	multiplier := node.NodeMultiplier

	// Apply plan base multiplier
	multiplier *= plan.BaseMultiplier

	// Get plan label multipliers
	labelMultipliers, err := s.planRepo.GetAllLabelMultipliers(plan.ID)
	if err != nil {
		return 0, err
	}

	// Apply label-specific multipliers for matching labels
	// Strategy: multiply all matching label multipliers
	for _, label := range node.Labels {
		if labelMult, exists := labelMultipliers[label.ID]; exists {
			multiplier *= labelMult
		}
	}

	return multiplier, nil
}

func (s *accountingService) GetCurrentUsage(userID uint64) (*models.UsagePeriod, error) {
	return s.usageRepo.GetCurrentPeriod(userID)
}

func (s *accountingService) CheckAndResetPeriods() error {
	// This is a background job that checks all users with current periods
	// and resets them if the period has ended

	// For simplicity, we'll implement a basic version
	// In production, you'd want to batch this and use more efficient queries

	s.logger.Info("Checking and resetting periods")

	// This would need to query all active periods and check if they need reset
	// For now, this is a placeholder that should be enhanced

	return nil
}

func (s *accountingService) InitializeUserPeriod(userID uint64) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.PlanID == nil {
		return errors.New("user has no plan")
	}

	plan, err := s.planRepo.FindByID(*user.PlanID)
	if err != nil {
		return err
	}

	now := time.Now()
	periodStart, periodEnd := s.calculatePeriodBounds(now, plan.ResetPeriod)

	period := &models.UsagePeriod{
		UserID:      user.ID,
		PlanID:      *user.PlanID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		IsCurrent:   true,
	}

	return s.usageRepo.CreatePeriod(period)
}

func (s *accountingService) calculatePeriodBounds(now time.Time, resetPeriod string) (time.Time, time.Time) {
	var start, end time.Time

	switch resetPeriod {
	case "daily":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 0, 1)
	case "weekly":
		weekday := int(now.Weekday())
		start = now.AddDate(0, 0, -weekday)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end = start.AddDate(0, 0, 7)
	case "monthly":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	case "yearly":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(1, 0, 0)
	default: // "none"
		start = now
		end = now.AddDate(100, 0, 0) // Far future
	}

	return start, end
}
