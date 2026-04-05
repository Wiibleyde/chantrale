package duty

import (
	"time"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

var botStartTime = time.Now()

// StartScheduler starts the daily 5:55 AM duty reset goroutine.
func StartScheduler(s *discordgo.Session) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 5, 55, 0, 0, now.Location())
			if !now.Before(next) {
				next = next.Add(24 * time.Hour)
			}
			logger.Info("Next duty reset scheduled", "at", next.Format(time.RFC3339))
			time.Sleep(time.Until(next))

			logger.Info("Running daily duty reset at 05:55")
			runReset(s)
			botStartTime = time.Now()
		}
	}()
}

// SendShutdownSummary sends a history summary to each logs channel. Called on bot shutdown.
func SendShutdownSummary(s *discordgo.Session) {
	logger.Info("Sending duty shutdown summary...")
	sendSummaryToAll(s, false)
}

// runReset sends summary, resets history, and strips all duty/oncall/offRadio roles.
func runReset(s *discordgo.Session) {
	sendSummaryToAll(s, true)
}

func sendSummaryToAll(s *discordgo.Session, stripRoles bool) {
	var dms []models.DutyManager
	if err := database.DB.Find(&dms).Error; err != nil {
		logger.Error("Error fetching DutyManagers for summary", "error", err)
		return
	}

	for _, dm := range dms {
		onDuty, onCall, offRadio := popHistory(dm.GuildID)

		if dm.LogsChannelID != nil {
			embed := BuildSummaryEmbed(botStartTime, time.Now(), onDuty, onCall, offRadio)
			if _, err := s.ChannelMessageSendComplex(*dm.LogsChannelID, &discordgo.MessageSend{
				Embeds: []*discordgo.MessageEmbed{embed},
			}); err != nil {
				logger.Error("Error sending duty summary", "guild", dm.GuildID, "error", err)
			}
		}

		if stripRoles {
			stripDutyRoles(s, dm)
		}
	}
}

func stripDutyRoles(s *discordgo.Session, dm models.DutyManager) {
	members, err := s.GuildMembers(dm.GuildID, "", 1000)
	if err != nil {
		logger.Error("Error fetching members for role strip", "guild", dm.GuildID, "error", err)
		return
	}

	for _, member := range members {
		for _, roleID := range member.Roles {
			if (dm.DutyRoleID != nil && roleID == *dm.DutyRoleID) ||
				(dm.OnCallRoleID != nil && roleID == *dm.OnCallRoleID) ||
				(dm.OffRadioRoleID != nil && roleID == *dm.OffRadioRoleID) {
				if err := s.GuildMemberRoleRemove(dm.GuildID, member.User.ID, roleID); err != nil {
					logger.Error("Error removing role during reset", "user", member.User.ID, "role", roleID, "error", err)
				}
			}
		}
	}
}
