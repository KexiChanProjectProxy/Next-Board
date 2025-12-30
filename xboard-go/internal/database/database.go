package database

import (
	"fmt"
	"time"

	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/config"
	"github.com/KexiChanProjectProxy/Next-Board/xboard-go/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Label{},
		&models.Plan{},
		&models.PlanLabel{},
		&models.PlanLabelMultiplier{},
		&models.Node{},
		&models.NodeLabel{},
		&models.UsagePeriod{},
		&models.NodeUsage{},
		&models.TelegramThreshold{},
		&models.UserUUID{},
		&models.OnlineUser{},
		&models.RefreshToken{},
	)
}
