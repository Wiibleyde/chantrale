package duty

import (
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

func HandleDutyButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handleRoleToggle(s, i, "duty")
}

func HandleOnCallButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handleRoleToggle(s, i, "oncall")
}

func HandleOffRadioButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handleRoleToggle(s, i, "offradio")
}

func handleRoleToggle(s *discordgo.Session, i *discordgo.InteractionCreate, roleType string) {
	var dm models.DutyManager
	if err := database.DB.Where("guild_id = ? AND message_id = ?", i.GuildID, i.Message.ID).First(&dm).Error; err != nil {
		respondEphemeral(s, i, "Gestionnaire de service introuvable.")
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
		respondEphemeral(s, i, "Rôle non configuré.")
		return
	}

	userID := i.Member.User.ID
	hasRole := false
	hasOpposite := false
	for _, r := range i.Member.Roles {
		if r == *roleID {
			hasRole = true
		}
		if oppositeRoleID != nil && r == *oppositeRoleID {
			hasOpposite = true
		}
	}

	var msgContent string
	if hasRole {
		if err := s.GuildMemberRoleRemove(i.GuildID, userID, *roleID); err != nil {
			logger.Error("Error removing role", "error", err)
			respondEphemeral(s, i, "Erreur lors de la modification du rôle.")
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
		// Remove the opposite role before adding this one (duty <-> oncall swap)
		if hasOpposite && oppositeRoleID != nil {
			if err := s.GuildMemberRoleRemove(i.GuildID, userID, *oppositeRoleID); err != nil {
				logger.Error("Error removing opposite role", "error", err)
				respondEphemeral(s, i, "Erreur lors de la modification du rôle.")
				return
			}
		}
		if err := s.GuildMemberRoleAdd(i.GuildID, userID, *roleID); err != nil {
			logger.Error("Error adding role", "error", err)
			respondEphemeral(s, i, "Erreur lors de la modification du rôle.")
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

	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		logger.Error("Error fetching members", "error", err)
	}

	var onDuty, onCall, offRadio []string
	if dm.DutyRoleID != nil {
		onDuty = membersWithRole(members, *dm.DutyRoleID)
	}
	if dm.OnCallRoleID != nil {
		onCall = membersWithRole(members, *dm.OnCallRoleID)
	}
	if dm.OffRadioRoleID != nil {
		offRadio = membersWithRole(members, *dm.OffRadioRoleID)
	}

	embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)

	components := []discordgo.MessageComponent{row}
	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         i.Message.ID,
		Channel:    dm.ChannelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		logger.Error("Error editing duty message", "error", err)
	}

	respondEphemeral(s, i, msgContent)
}
