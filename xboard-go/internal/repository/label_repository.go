package repository

import (
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type LabelRepository interface {
	Create(label *models.Label) error
	FindByID(id uint64) (*models.Label, error)
	FindByName(name string) (*models.Label, error)
	Update(label *models.Label) error
	Delete(id uint64) error
	List(offset, limit int) ([]models.Label, int64, error)
	FindAll() ([]models.Label, error)
}

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) LabelRepository {
	return &labelRepository{db: db}
}

func (r *labelRepository) Create(label *models.Label) error {
	return r.db.Create(label).Error
}

func (r *labelRepository) FindByID(id uint64) (*models.Label, error) {
	var label models.Label
	err := r.db.First(&label, id).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) FindByName(name string) (*models.Label, error) {
	var label models.Label
	err := r.db.Where("name = ?", name).First(&label).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) Update(label *models.Label) error {
	return r.db.Save(label).Error
}

func (r *labelRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Label{}, id).Error
}

func (r *labelRepository) List(offset, limit int) ([]models.Label, int64, error) {
	var labels []models.Label
	var total int64

	if err := r.db.Model(&models.Label{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(limit).Find(&labels).Error
	return labels, total, err
}

func (r *labelRepository) FindAll() ([]models.Label, error) {
	var labels []models.Label
	err := r.db.Find(&labels).Error
	return labels, err
}
