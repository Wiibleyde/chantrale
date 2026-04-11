package database

import (
	"fmt"

	"LsmsBot/internal/config"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	cfg := config.Load()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	if err := db.AutoMigrate(&models.DutyManager{}, &models.BedManager{}, &models.BedAssignment{}, &stats.StatEvent{}); err != nil {
		logger.Fatal("Failed to migrate database", "error", err)
	}
	DB = db
	logger.Info("Database connected and migrated")
	stats.Init(db)
}
