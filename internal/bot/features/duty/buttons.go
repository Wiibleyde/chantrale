package duty

import (
	"LsmsBot/internal/bot/helpers"
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func HandleDutyButton(e *events.ComponentInteractionCreate) {
	handleRoleToggle(e, "duty")
}

func HandleOnCallButton(e *events.ComponentInteractionCreate) {
	handleRoleToggle(e, "oncall")
}

func HandleOffRadioButton(e *events.ComponentInteractionCreate) {
	handleRoleToggle(e, "offradio")
}

func handleRoleToggle(e *events.ComponentInteractionCreate, roleType string) {
	guildID := *e.GuildID()
	msgID := e.Message.ID.String()

	var dm models.DutyManager
	if err := database.DB.Where("guild_id = ? AND message_id = ?", guildID.String(), msgID).First(&dm).Error; err != nil {
		helpers.RespondEphemeral(e, "Gestionnaire de service introuvable.")
		return
	}

	var roleID *string
	var oppositeRoleID *string
	switch roleType {
	case "duty":
		roleID = dm.DutyRoleID
		oppositeRoleID = dm.OnCallRoleID
	case "oncall":
		roleID = dm.OnCallRoleID
		oppositeRoleID = dm.DutyRoleID
	case "offradio":
		roleID = dm.OffRadioRoleID
	}

	if roleID == nil {
		helpers.RespondEphemeral(e, "Rôle non configuré.")
		return
	}

	member := e.Member()
	if member == nil {
		helpers.RespondEphemeral(e, "Erreur: membre introuvable.")
		return
	}
	userID := member.User.ID

	roleSnowflake, err := snowflake.Parse(*roleID)
	if err != nil {
		helpers.RespondEphemeral(e, "Erreur: rôle invalide.")
		return
	}

	hasRole := false
	hasOpposite := false
	for _, r := range member.RoleIDs {
		if r == roleSnowflake {
			hasRole = true
		}
		if oppositeRoleID != nil {
			oppSnowflake, err := snowflake.Parse(*oppositeRoleID)
			if err == nil && r == oppSnowflake {
				hasOpposite = true
			}
		}
	}

	client := e.Client()
	var msgContent string
	if hasRole {
		if err := client.Rest.RemoveMemberRole(guildID, userID, roleSnowflake); err != nil {
			logger.Error("Error removing role", "error", err)
			helpers.RespondEphemeral(e, "Erreur lors de la modification du rôle.")
			return
		}
		switch roleType {
		case "duty":
			msgContent = "Vous avez quitté le service."
		case "oncall":
			msgContent = "Vous avez quitté le semi service."
		case "offradio":
			msgContent = "Vous avez quitté le off radio."
		}
	} else {
		if hasOpposite && oppositeRoleID != nil {
			oppSnowflake, err := snowflake.Parse(*oppositeRoleID)
			if err == nil {
				if err := client.Rest.RemoveMemberRole(guildID, userID, oppSnowflake); err != nil {
					logger.Error("Error removing opposite role", "error", err)
					helpers.RespondEphemeral(e, "Erreur lors de la modification du rôle.")
					return
				}
			}
		}
		if err := client.Rest.AddMemberRole(guildID, userID, roleSnowflake); err != nil {
			logger.Error("Error adding role", "error", err)
			helpers.RespondEphemeral(e, "Erreur lors de la modification du rôle.")
			return
		}
		switch roleType {
		case "duty":
			msgContent = "Vous avez pris le service."
		case "oncall":
			msgContent = "Vous avez pris le semi service."
		case "offradio":
			msgContent = "Vous êtes passé off radio."
		}
	}

	helpers.RespondEphemeral(e, msgContent)

	// Send log message and update the embed directly, since GuildMemberUpdate
	// may not have OldMember cached and would miss this change.
	displayName := memberDisplayName(member.Member)
	take := !hasRole

	action := "duty.role_leave"
	if take {
		action = "duty.role_take"
	}
	stats.Record(guildID.String(), userID.String(), action, map[string]any{
		"role_type":    roleType,
		"display_name": displayName,
	})

	if dm.LogsChannelID != nil {
		if logsChannelID, err := snowflake.Parse(*dm.LogsChannelID); err == nil {
			var comps []discord.LayoutComponent
			switch roleType {
			case "duty":
				comps = BuildDutyUpdateComponents(displayName, take)
				if take {
					trackDuty(guildID.String(), displayName)
				}
			case "oncall":
				comps = BuildOnCallUpdateComponents(displayName, take)
				if take {
					trackOnCall(guildID.String(), displayName)
				}
			case "offradio":
				comps = BuildOffRadioUpdateComponents(displayName, take)
				if take {
					trackOffRadio(guildID.String(), displayName)
				}
			}
			if comps != nil {
				if _, err := client.Rest.CreateMessage(logsChannelID, discord.NewMessageCreateV2(comps...)); err != nil {
					logger.Error("Error sending log embed", "error", err)
				}
			}
		}
	}

	scheduleEmbedUpdate(client, guildID, dm)
}

