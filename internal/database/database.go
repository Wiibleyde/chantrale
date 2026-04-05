package database

import (
	"fmt"
	"log"

	"LsmsBot/internal/config"
	"LsmsBot/internal/database/models"

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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.AutoMigrate(&models.DutyManager{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	DB = db
	log.Println("Database connected and migrated.")
}
