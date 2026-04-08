package duty

import (
	"time"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

var botStartTime = time.Now()

func StartScheduler(client *bot.Client) {
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
			runReset(client)
			botStartTime = time.Now()
		}
	}()
}

func SendShutdownSummary(client *bot.Client) {
	logger.Info("Sending duty shutdown summary...")
	sendSummaryToAll(client, false)
}

func runReset(client *bot.Client) {
	sendSummaryToAll(client, true)
}

func sendSummaryToAll(client *bot.Client, stripRoles bool) {
	var dms []models.DutyManager
	if err := database.DB.Find(&dms).Error; err != nil {
		logger.Error("Error fetching DutyManagers for summary", "error", err)
		return
	}

	for _, dm := range dms {
		onDuty, onCall, offRadio := popHistory(dm.GuildID)

		if dm.LogsChannelID != nil {
			logsChannelID, err := snowflake.Parse(*dm.LogsChannelID)
			if err == nil {
				embed := BuildSummaryEmbed(botStartTime, time.Now(), onDuty, onCall, offRadio)
				if _, err := client.Rest.CreateMessage(logsChannelID, discord.MessageCreate{
					Embeds: []discord.Embed{embed},
				}); err != nil {
					logger.Error("Error sending duty summary", "guild", dm.GuildID, "error", err)
				}
			}
		}

		if stripRoles {
			stripDutyRoles(client, dm)
			setGuildCounts(dm.GuildID, 0, 0)
		}
	}

	if stripRoles {
		updateBotPresence(client)
	}
}

func stripDutyRoles(client *bot.Client, dm models.DutyManager) {
	guildID, err := snowflake.Parse(dm.GuildID)
	if err != nil {
		return
	}

	members, err := client.Rest.GetMembers(guildID, 1000, 0)
	if err != nil {
		logger.Error("Error fetching members for role strip", "guild", dm.GuildID, "error", err)
		return
	}

	for _, member := range members {
		for _, roleID := range member.RoleIDs {
			roleIDStr := roleID.String()
			if (dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr) ||
				(dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr) ||
				(dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr) {
				if err := client.Rest.RemoveMemberRole(guildID, member.User.ID, roleID); err != nil {
					logger.Error("Error removing role during reset", "user", member.User.ID, "role", roleID, "error", err)
				}
			}
		}
	}
}
