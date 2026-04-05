package radio

import (
	"log"

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
		log.Printf("Error sending radio message: %v", err)
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
		log.Printf("Error responding to interaction: %v", err)
	}
}
