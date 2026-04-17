package models

type MortuaryManager struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	GuildID   string `gorm:"uniqueIndex"`
	ChannelID string
	MessageID string
}

type MortuaryAssignment struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	GuildID      string  `gorm:"index"`
	LockerNumber int
	Name         string
	Comment      *string
}
