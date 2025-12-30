package repository

import (
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type UUIDRepository interface {
	Create(userUUID *models.UserUUID) error
	FindByUUID(uuid string) (*models.UserUUID, error)
	FindByUserID(userID uint64) (*models.UserUUID, error)
	GetAllUserUUIDs() (map[uint64]string, error)
}

type uuidRepository struct {
	db *gorm.DB
}

func NewUUIDRepository(db *gorm.DB) UUIDRepository {
	return &uuidRepository{db: db}
}

func (r *uuidRepository) Create(userUUID *models.UserUUID) error {
	return r.db.Create(userUUID).Error
}

func (r *uuidRepository) FindByUUID(uuid string) (*models.UserUUID, error) {
	var userUUID models.UserUUID
	err := r.db.Where("uuid = ?", uuid).First(&userUUID).Error
	if err != nil {
		return nil, err
	}
	return &userUUID, nil
}

func (r *uuidRepository) FindByUserID(userID uint64) (*models.UserUUID, error) {
	var userUUID models.UserUUID
	err := r.db.Where("user_id = ?", userID).First(&userUUID).Error
	if err != nil {
		return nil, err
	}
	return &userUUID, nil
}

func (r *uuidRepository) GetAllUserUUIDs() (map[uint64]string, error) {
	var uuids []models.UserUUID
	err := r.db.Find(&uuids).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint64]string)
	for _, uuid := range uuids {
		result[uuid.UserID] = uuid.UUID
	}
	return result, nil
}
