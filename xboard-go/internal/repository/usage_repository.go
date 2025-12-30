package repository

import (
	"time"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type UsageRepository interface {
	GetCurrentPeriod(userID uint64) (*models.UsagePeriod, error)
	CreatePeriod(period *models.UsagePeriod) error
	UpdatePeriod(period *models.UsagePeriod) error
	ClosePeriod(periodID uint64) error
	GetPeriodHistory(userID uint64, start, end time.Time) ([]models.UsagePeriod, error)
	GetNodeUsage(periodID uint64) ([]models.NodeUsage, error)
	GetNodeUsageByUserAndNode(userID, nodeID, periodID uint64) (*models.NodeUsage, error)
	CreateNodeUsage(usage *models.NodeUsage) error
	UpdateNodeUsage(usage *models.NodeUsage) error
	IncrementUsage(userID, nodeID uint64, realUp, realDown, billableUp, billableDown uint64) error
}

type usageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) UsageRepository {
	return &usageRepository{db: db}
}

func (r *usageRepository) GetCurrentPeriod(userID uint64) (*models.UsagePeriod, error) {
	var period models.UsagePeriod
	err := r.db.Where("user_id = ? AND is_current = ?", userID, true).First(&period).Error
	if err != nil {
		return nil, err
	}
	return &period, nil
}

func (r *usageRepository) CreatePeriod(period *models.UsagePeriod) error {
	return r.db.Create(period).Error
}

func (r *usageRepository) UpdatePeriod(period *models.UsagePeriod) error {
	return r.db.Save(period).Error
}

func (r *usageRepository) ClosePeriod(periodID uint64) error {
	return r.db.Model(&models.UsagePeriod{}).Where("id = ?", periodID).Update("is_current", false).Error
}

func (r *usageRepository) GetPeriodHistory(userID uint64, start, end time.Time) ([]models.UsagePeriod, error) {
	var periods []models.UsagePeriod
	err := r.db.Where("user_id = ? AND period_start >= ? AND period_end <= ?", userID, start, end).
		Order("period_start DESC").
		Find(&periods).Error
	return periods, err
}

func (r *usageRepository) GetNodeUsage(periodID uint64) ([]models.NodeUsage, error) {
	var usages []models.NodeUsage
	err := r.db.Where("period_id = ?", periodID).Find(&usages).Error
	return usages, err
}

func (r *usageRepository) GetNodeUsageByUserAndNode(userID, nodeID, periodID uint64) (*models.NodeUsage, error) {
	var usage models.NodeUsage
	err := r.db.Where("user_id = ? AND node_id = ? AND period_id = ?", userID, nodeID, periodID).First(&usage).Error
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *usageRepository) CreateNodeUsage(usage *models.NodeUsage) error {
	return r.db.Create(usage).Error
}

func (r *usageRepository) UpdateNodeUsage(usage *models.NodeUsage) error {
	return r.db.Save(usage).Error
}

func (r *usageRepository) IncrementUsage(userID, nodeID uint64, realUp, realDown, billableUp, billableDown uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get current period
		var period models.UsagePeriod
		if err := tx.Where("user_id = ? AND is_current = ?", userID, true).First(&period).Error; err != nil {
			return err
		}

		// Update period totals
		if err := tx.Model(&period).Updates(map[string]interface{}{
			"real_bytes_up":       gorm.Expr("real_bytes_up + ?", realUp),
			"real_bytes_down":     gorm.Expr("real_bytes_down + ?", realDown),
			"billable_bytes_up":   gorm.Expr("billable_bytes_up + ?", billableUp),
			"billable_bytes_down": gorm.Expr("billable_bytes_down + ?", billableDown),
		}).Error; err != nil {
			return err
		}

		// Update or create node usage
		var nodeUsage models.NodeUsage
		result := tx.Where("user_id = ? AND node_id = ? AND period_id = ?", userID, nodeID, period.ID).First(&nodeUsage)

		if result.Error == gorm.ErrRecordNotFound {
			nodeUsage = models.NodeUsage{
				UserID:            userID,
				NodeID:            nodeID,
				PeriodID:          period.ID,
				RealBytesUp:       realUp,
				RealBytesDown:     realDown,
				BillableBytesUp:   billableUp,
				BillableBytesDown: billableDown,
			}
			return tx.Create(&nodeUsage).Error
		}

		if result.Error != nil {
			return result.Error
		}

		return tx.Model(&nodeUsage).Updates(map[string]interface{}{
			"real_bytes_up":       gorm.Expr("real_bytes_up + ?", realUp),
			"real_bytes_down":     gorm.Expr("real_bytes_down + ?", realDown),
			"billable_bytes_up":   gorm.Expr("billable_bytes_up + ?", billableUp),
			"billable_bytes_down": gorm.Expr("billable_bytes_down + ?", billableDown),
		}).Error
	})
}
