package models

type BedManager struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	GuildID   string `gorm:"uniqueIndex"`
	ChannelID string
	MessageID string
}

type BedAssignment struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	GuildID     string `gorm:"index"`
	BedLetter   string
	Name        string
	UnderArrest bool
	Death       bool
}
