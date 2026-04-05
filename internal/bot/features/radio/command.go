package radio

import (
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "radio",
		Description: "Créer un gestionnaire de radios",
	},
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionManageChannels == 0 {
		respondEphemeral(s, i, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	embed := BuildRadioEmbed(nil)
	components := BuildRadioComponents(nil)

	if _, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	}); err != nil {
		logger.Error("Error sending radio message", "error", err)
		respondEphemeral(s, i, "Erreur lors de l'envoi du message.")
		return
	}

	respondEphemeral(s, i, "Gestionnaire de radios créé avec succès.")
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}
