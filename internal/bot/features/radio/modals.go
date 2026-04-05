package radio

import (
	"encoding/base64"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HandleRadioAddModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID
	parts := strings.SplitN(customID, "--", 3)
	if len(parts) < 3 {
		respondEphemeral(s, i, "Erreur: identifiant invalide.")
		return
	}

	channelID := parts[1]
	messageID := parts[2]

	data := i.ModalSubmitData()
	name := modalFieldValue(data, "name")
	frequency := modalFieldValue(data, "frequency")

	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		log.Printf("Error fetching message: %v", err)
		respondEphemeral(s, i, "Erreur lors de la récupération du message.")
		return
	}

	radios := ParseRadiosFromEmbed(msg.Embeds)
	radios = append(radios, RadioEntry{Name: name, Frequency: frequency})

	embed := BuildRadioEmbed(radios)
	components := BuildRadioComponents(radios)

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("Error editing radio message: %v", err)
		respondEphemeral(s, i, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(s, i, "Radio ajoutée avec succès.")
}

func HandleRadioEditModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.ModalSubmitData().CustomID
	parts := strings.SplitN(customID, "--", 4)
	if len(parts) < 4 {
		respondEphemeral(s, i, "Erreur: identifiant invalide.")
		return
	}

	channelID := parts[1]
	messageID := parts[2]
	encodedName := parts[3]

	nameBytes, err := base64.RawURLEncoding.DecodeString(encodedName)
	if err != nil {
		respondEphemeral(s, i, "Erreur: nom de radio invalide.")
		return
	}
	originalName := string(nameBytes)

	data := i.ModalSubmitData()
	newName := modalFieldValue(data, "name")
	newFreq := modalFieldValue(data, "frequency")

	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		log.Printf("Error fetching message: %v", err)
		respondEphemeral(s, i, "Erreur lors de la récupération du message.")
		return
	}

	radios := ParseRadiosFromEmbed(msg.Embeds)
	for idx, r := range radios {
		if r.Name == originalName {
			radios[idx] = RadioEntry{Name: newName, Frequency: newFreq}
			break
		}
	}

	embed := BuildRadioEmbed(radios)
	components := BuildRadioComponents(radios)

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("Error editing radio message: %v", err)
		respondEphemeral(s, i, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(s, i, "Radio modifiée avec succès.")
}

func modalFieldValue(data discordgo.ModalSubmitInteractionData, customID string) string {
	for _, row := range data.Components {
		actionsRow, ok := row.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, comp := range actionsRow.Components {
			input, ok := comp.(*discordgo.TextInput)
			if !ok {
				continue
			}
			if input.CustomID == customID {
				return input.Value
			}
		}
	}
	return ""
}
