package labo

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func HandleCancelButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	messageID := i.Message.ID

	found, entry := Queue.CancelByMessageID(messageID)
	if !found {
		respondEphemeral(s, i, "Analyse introuvable ou déjà terminée.")
		return
	}

	if entry.UserID != i.Member.User.ID {
		Queue.Add(entry)
		respondEphemeral(s, i, "Vous ne pouvez pas annuler l'analyse d'un autre membre.")
		return
	}

	embed := BuildLaboCancelledEmbed(entry)
	components := []discordgo.MessageComponent{}

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    i.ChannelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("Error editing labo message: %v", err)
	}

	respondEphemeral(s, i, "Analyse annulée avec succès.")
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
