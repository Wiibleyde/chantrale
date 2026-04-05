package models

type DutyManager struct {
	ID             uint    `gorm:"primaryKey;autoIncrement"`
	GuildID        string  `gorm:"index"`
	ChannelID      string
	MessageID      *string
	DutyRoleID     *string
	OnCallRoleID   *string
	OffRadioRoleID *string
	LogsChannelID  *string
}
