package models

import (
	"time"

	"gorm.io/datatypes"
)

type StatEvent struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	GuildID   string         `gorm:"index;not null;index:idx_guild_type_time"`
	UserID    string         `gorm:"index"`
	EventType string         `gorm:"index;not null;index:idx_guild_type_time"`
	Payload   datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"index;autoCreateTime;index:idx_guild_type_time"`
}
