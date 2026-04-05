package radio

import (
	"encoding/base64"
	"strings"

	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

func boolPtr(b bool) *bool { return &b }

func HandleRadioAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID
	messageID := i.Message.ID

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "lsmsRadioAddModal--" + channelID + "--" + messageID,
			Title:    "Ajouter une radio",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "name",
							Label:       "Nom de la radio",
							Style:       discordgo.TextInputShort,
							Required:    boolPtr(true),
							Placeholder: "Ex: Radio LSMS",
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "frequency",
							Label:       "Fréquence",
							Style:       discordgo.TextInputShort,
							Required:    boolPtr(true),
							Placeholder: "Ex: 101.5",
						},
					},
				},
			},
		},
	}); err != nil {
		logger.Error("Error responding with modal", "error", err)
	}
}

func HandleRadioRemove(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channelID := i.ChannelID
	messageID := i.Message.ID

	radios := ParseRadiosFromEmbed(i.Message.Embeds)
	if len(radios) == 0 {
		respondEphemeral(s, i, "Aucune radio à supprimer.")
		return
	}

	var options []discordgo.SelectMenuOption
	for _, r := range radios {
		options = append(options, discordgo.SelectMenuOption{
			Label: r.Name,
			Value: r.Name,
		})
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "lsmsRadioRemoveSelect--" + channelID + "--" + messageID,
							Placeholder: "Sélectionner une radio à supprimer",
							Options:     options,
						},
					},
				},
			},
		},
	}); err != nil {
		logger.Error("Error responding with select menu", "error", err)
	}
}

func HandleRadioEdit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	parts := strings.SplitN(customID, "--", 2)
	if len(parts) < 2 {
		respondEphemeral(s, i, "Erreur: identifiant invalide.")
		return
	}

	encodedName := parts[1]
	nameBytes, err := base64.RawURLEncoding.DecodeString(encodedName)
	if err != nil {
		respondEphemeral(s, i, "Erreur: nom de radio invalide.")
		return
	}
	radioName := string(nameBytes)

	channelID := i.ChannelID
	messageID := i.Message.ID

	radios := ParseRadiosFromEmbed(i.Message.Embeds)
	currentFreq := ""
	for _, r := range radios {
		if r.Name == radioName {
			currentFreq = r.Frequency
			break
		}
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte(radioName))

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "lsmsRadioEditModal--" + channelID + "--" + messageID + "--" + encoded,
			Title:    "Modifier la radio " + radioName,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "name",
							Label:    "Nom de la radio",
							Style:    discordgo.TextInputShort,
							Required: boolPtr(true),
							Value:    radioName,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "frequency",
							Label:    "Fréquence",
							Style:    discordgo.TextInputShort,
							Required: boolPtr(true),
							Value:    currentFreq,
						},
					},
				},
			},
		},
	}); err != nil {
		logger.Error("Error responding with modal", "error", err)
	}
}

func HandleRadioRemoveSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	parts := strings.SplitN(customID, "--", 3)
	if len(parts) < 3 {
		respondEphemeral(s, i, "Erreur: identifiant invalide.")
		return
	}

	channelID := parts[1]
	messageID := parts[2]

	values := i.MessageComponentData().Values
	if len(values) == 0 {
		respondEphemeral(s, i, "Aucune radio sélectionnée.")
		return
	}
	selectedName := values[0]

	msg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		logger.Error("Error fetching message", "error", err)
		respondEphemeral(s, i, "Erreur lors de la récupération du message.")
		return
	}

	radios := ParseRadiosFromEmbed(msg.Embeds)
	var newRadios []RadioEntry
	for _, r := range radios {
		if r.Name != selectedName {
			newRadios = append(newRadios, r)
		}
	}

	embed := BuildRadioEmbed(newRadios)
	components := BuildRadioComponents(newRadios)

	if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		logger.Error("Error editing radio message", "error", err)
		respondEphemeral(s, i, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(s, i, "Radio supprimée avec succès.")
}
