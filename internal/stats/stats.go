package stats

import (
	"encoding/json"
	"time"

	"LsmsBot/internal/logger"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StatEvent struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	GuildID   string         `gorm:"index;not null;index:idx_guild_type_time"`
	UserID    string         `gorm:"index"`
	EventType string         `gorm:"index;not null;index:idx_guild_type_time"`
	Payload   datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"index;autoCreateTime;index:idx_guild_type_time"`
}

const bufferSize = 256

var (
	eventCh chan *StatEvent
	db      *gorm.DB
)

func Init(gormDB *gorm.DB) {
	db = gormDB
	eventCh = make(chan *StatEvent, bufferSize)
	go writer()
}

// Record enqueues a stat event for asynchronous persistence.
// It never blocks the caller: if the buffer is full the event is dropped.
func Record(guildID, userID, eventType string, payload map[string]any) {
	if eventCh == nil {
		return
	}

	var raw datatypes.JSON
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal stat payload", "error", err, "eventType", eventType)
			return
		}
		raw = b
	}

	ev := &StatEvent{
		GuildID:   guildID,
		UserID:    userID,
		EventType: eventType,
		Payload:   raw,
	}

	select {
	case eventCh <- ev:
	default:
		logger.Warn("Stats buffer full, dropping event", "eventType", eventType)
	}
}

func writer() {
	for ev := range eventCh {
		if err := db.Create(ev).Error; err != nil {
			logger.Error("Failed to persist stat event", "error", err, "eventType", ev.EventType)
		}
	}
}
