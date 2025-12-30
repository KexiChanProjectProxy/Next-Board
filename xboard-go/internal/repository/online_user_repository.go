package repository

import (
	"time"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/gorm"
)

type OnlineUserRepository interface {
	UpsertOnlineUser(userID, nodeID uint64, ipAddress string) error
	GetOnlineDeviceCount(userID uint64) (uint, error)
	GetAllOnlineDeviceCounts() (map[uint64]uint, error)
	CleanupStaleOnlineUsers(before time.Time) error
	DeleteByUser(userID uint64) error
}

type onlineUserRepository struct {
	db *gorm.DB
}

func NewOnlineUserRepository(db *gorm.DB) OnlineUserRepository {
	return &onlineUserRepository{db: db}
}

func (r *onlineUserRepository) UpsertOnlineUser(userID, nodeID uint64, ipAddress string) error {
	var onlineUser models.OnlineUser
	result := r.db.Where("user_id = ? AND node_id = ? AND ip_address = ?", userID, nodeID, ipAddress).First(&onlineUser)

	if result.Error == gorm.ErrRecordNotFound {
		onlineUser = models.OnlineUser{
			UserID:     userID,
			NodeID:     nodeID,
			IPAddress:  ipAddress,
			LastSeenAt: time.Now(),
		}
		return r.db.Create(&onlineUser).Error
	}

	if result.Error != nil {
		return result.Error
	}

	onlineUser.LastSeenAt = time.Now()
	return r.db.Save(&onlineUser).Error
}

func (r *onlineUserRepository) GetOnlineDeviceCount(userID uint64) (uint, error) {
	var count int64
	err := r.db.Model(&models.OnlineUser{}).
		Where("user_id = ?", userID).
		Distinct("ip_address").
		Count(&count).Error
	return uint(count), err
}

func (r *onlineUserRepository) GetAllOnlineDeviceCounts() (map[uint64]uint, error) {
	var results []struct {
		UserID uint64
		Count  uint
	}

	err := r.db.Model(&models.OnlineUser{}).
		Select("user_id, COUNT(DISTINCT ip_address) as count").
		Group("user_id").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[uint64]uint)
	for _, r := range results {
		counts[r.UserID] = r.Count
	}

	return counts, nil
}

func (r *onlineUserRepository) CleanupStaleOnlineUsers(before time.Time) error {
	return r.db.Where("last_seen_at < ?", before).Delete(&models.OnlineUser{}).Error
}

func (r *onlineUserRepository) DeleteByUser(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.OnlineUser{}).Error
}
