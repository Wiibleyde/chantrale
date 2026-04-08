package radio

import (
	"encoding/base64"
	"strings"

	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func HandleRadioAddModal(e *events.ModalSubmitInteractionCreate) {
	customID := e.Data.CustomID
	parts := strings.SplitN(customID, "--", 3)
	if len(parts) < 3 {
		respondEphemeral(e, "Erreur: identifiant invalide.")
		return
	}

	channelIDStr := parts[1]
	messageIDStr := parts[2]

	name := e.Data.Text("name")
	frequency := e.Data.Text("frequency")

	chanID := snowflake.MustParse(channelIDStr)
	msgID := snowflake.MustParse(messageIDStr)

	msg, err := e.Client().Rest.GetMessage(chanID, msgID)
	if err != nil {
		logger.Error("Error fetching message", "error", err)
		respondEphemeral(e, "Erreur lors de la récupération du message.")
		return
	}

	radios := ParseRadiosFromEmbed(msg.Embeds)
	radios = append(radios, RadioEntry{Name: name, Frequency: frequency})

	embed := BuildRadioEmbed(radios)
	components := BuildRadioComponents(radios)
	embeds := []discord.Embed{embed}
	if _, err := e.Client().Rest.UpdateMessage(chanID, msgID, discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	}); err != nil {
		logger.Error("Error editing radio message", "error", err)
		respondEphemeral(e, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(e, "Radio ajoutée avec succès.")
}

func HandleRadioEditModal(e *events.ModalSubmitInteractionCreate) {
	customID := e.Data.CustomID
	parts := strings.SplitN(customID, "--", 4)
	if len(parts) < 4 {
		respondEphemeral(e, "Erreur: identifiant invalide.")
		return
	}

	channelIDStr := parts[1]
	messageIDStr := parts[2]
	encodedName := parts[3]

	nameBytes, err := base64.RawURLEncoding.DecodeString(encodedName)
	if err != nil {
		respondEphemeral(e, "Erreur: nom de radio invalide.")
		return
	}
	originalName := string(nameBytes)

	newName := e.Data.Text("name")
	newFreq := e.Data.Text("frequency")

	chanID := snowflake.MustParse(channelIDStr)
	msgID := snowflake.MustParse(messageIDStr)

	msg, err := e.Client().Rest.GetMessage(chanID, msgID)
	if err != nil {
		logger.Error("Error fetching message", "error", err)
		respondEphemeral(e, "Erreur lors de la récupération du message.")
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
	embeds := []discord.Embed{embed}
	if _, err := e.Client().Rest.UpdateMessage(chanID, msgID, discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	}); err != nil {
		logger.Error("Error editing radio message", "error", err)
		respondEphemeral(e, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(e, "Radio modifiée avec succès.")
}
