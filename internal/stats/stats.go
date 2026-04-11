package stats

import (
	"encoding/json"

	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const bufferSize = 256

var (
	eventCh chan *models.StatEvent
	db      *gorm.DB
)

func Init(gormDB *gorm.DB) {
	db = gormDB
	eventCh = make(chan *models.StatEvent, bufferSize)
	go writer()
}

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

	ev := &models.StatEvent{
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
