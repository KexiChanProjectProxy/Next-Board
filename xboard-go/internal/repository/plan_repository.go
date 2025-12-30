package repository

import (
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(plan *models.Plan) error
	FindByID(id uint64) (*models.Plan, error)
	FindByIDWithLabels(id uint64) (*models.Plan, error)
	Update(plan *models.Plan) error
	Delete(id uint64) error
	List(offset, limit int) ([]models.Plan, int64, error)
	AddLabel(planID, labelID uint64) error
	RemoveLabel(planID, labelID uint64) error
	GetLabels(planID uint64) ([]models.Label, error)
	SetLabelMultiplier(planID, labelID uint64, multiplier float64) error
	GetLabelMultiplier(planID, labelID uint64) (float64, error)
	GetAllLabelMultipliers(planID uint64) (map[uint64]float64, error)
}

type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(plan *models.Plan) error {
	return r.db.Create(plan).Error
}

func (r *planRepository) FindByID(id uint64) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) FindByIDWithLabels(id uint64) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.Preload("Labels").First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) Update(plan *models.Plan) error {
	return r.db.Save(plan).Error
}

func (r *planRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Plan{}, id).Error
}

func (r *planRepository) List(offset, limit int) ([]models.Plan, int64, error) {
	var plans []models.Plan
	var total int64

	if err := r.db.Model(&models.Plan{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Labels").Offset(offset).Limit(limit).Find(&plans).Error
	return plans, total, err
}

func (r *planRepository) AddLabel(planID, labelID uint64) error {
	planLabel := &models.PlanLabel{
		PlanID:  planID,
		LabelID: labelID,
	}
	return r.db.Create(planLabel).Error
}

func (r *planRepository) RemoveLabel(planID, labelID uint64) error {
	return r.db.Where("plan_id = ? AND label_id = ?", planID, labelID).Delete(&models.PlanLabel{}).Error
}

func (r *planRepository) GetLabels(planID uint64) ([]models.Label, error) {
	var plan models.Plan
	err := r.db.Preload("Labels").First(&plan, planID).Error
	if err != nil {
		return nil, err
	}
	return plan.Labels, nil
}

func (r *planRepository) SetLabelMultiplier(planID, labelID uint64, multiplier float64) error {
	var plm models.PlanLabelMultiplier
	result := r.db.Where("plan_id = ? AND label_id = ?", planID, labelID).First(&plm)

	if result.Error == gorm.ErrRecordNotFound {
		plm = models.PlanLabelMultiplier{
			PlanID:     planID,
			LabelID:    labelID,
			Multiplier: multiplier,
		}
		return r.db.Create(&plm).Error
	}

	if result.Error != nil {
		return result.Error
	}

	plm.Multiplier = multiplier
	return r.db.Save(&plm).Error
}

func (r *planRepository) GetLabelMultiplier(planID, labelID uint64) (float64, error) {
	var plm models.PlanLabelMultiplier
	err := r.db.Where("plan_id = ? AND label_id = ?", planID, labelID).First(&plm).Error
	if err == gorm.ErrRecordNotFound {
		return 1.0, nil
	}
	if err != nil {
		return 0, err
	}
	return plm.Multiplier, nil
}

func (r *planRepository) GetAllLabelMultipliers(planID uint64) (map[uint64]float64, error) {
	var multipliers []models.PlanLabelMultiplier
	err := r.db.Where("plan_id = ?", planID).Find(&multipliers).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]float64)
	for _, m := range multipliers {
		result[m.LabelID] = m.Multiplier
	}
	return result, nil
}
