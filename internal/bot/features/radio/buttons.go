package radio

import (
	"encoding/base64"
	"strings"

	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func HandleRadioAdd(e *events.ComponentInteractionCreate) {
	channelID := e.Channel().ID().String()
	messageID := e.Message.ID.String()

	modal := discord.NewModalCreate("lsmsRadioAddModal--"+channelID+"--"+messageID, "Ajouter une radio").
		AddLabel("Nom de la radio", discord.NewShortTextInput("name").WithRequired(true).WithPlaceholder("Ex: Radio LSMS")).
		AddLabel("Fréquence", discord.NewShortTextInput("frequency").WithRequired(true).WithPlaceholder("Ex: 101.5"))

	if err := e.Modal(modal); err != nil {
		logger.Error("Error responding with modal", "error", err)
	}
}

func HandleRadioRemove(e *events.ComponentInteractionCreate) {
	channelID := e.Channel().ID().String()
	messageID := e.Message.ID.String()

	radios := ParseRadiosFromComponents(e.Message.Components)
	if len(radios) == 0 {
		respondEphemeral(e, "Aucune radio à supprimer.")
		return
	}

	var options []discord.StringSelectMenuOption
	for _, r := range radios {
		options = append(options, discord.StringSelectMenuOption{Label: r.Name, Value: r.Name})
	}

	if err := e.CreateMessage(discord.MessageCreate{
		Flags: discord.MessageFlagEphemeral,
		Components: []discord.LayoutComponent{
			discord.ActionRowComponent{Components: []discord.InteractiveComponent{
				discord.StringSelectMenuComponent{
					CustomID:    "lsmsRadioRemoveSelect--" + channelID + "--" + messageID,
					Placeholder: "Sélectionner une radio à supprimer",
					Options:     options,
				},
			}},
		},
	}); err != nil {
		logger.Error("Error responding with select menu", "error", err)
	}
}

func HandleRadioEdit(e *events.ComponentInteractionCreate) {
	customID := e.Data.CustomID()
	parts := strings.SplitN(customID, "--", 2)
	if len(parts) < 2 {
		respondEphemeral(e, "Erreur: identifiant invalide.")
		return
	}

	encodedName := parts[1]
	nameBytes, err := base64.RawURLEncoding.DecodeString(encodedName)
	if err != nil {
		respondEphemeral(e, "Erreur: nom de radio invalide.")
		return
	}
	radioName := string(nameBytes)

	channelID := e.Channel().ID().String()
	messageID := e.Message.ID.String()

	radios := ParseRadiosFromComponents(e.Message.Components)
	currentFreq := ""
	for _, r := range radios {
		if r.Name == radioName {
			currentFreq = r.Frequency
			break
		}
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte(radioName))

	modal := discord.NewModalCreate("lsmsRadioEditModal--"+channelID+"--"+messageID+"--"+encoded, "Modifier la radio "+radioName).
		AddLabel("Nom de la radio", discord.NewShortTextInput("name").WithRequired(true).WithValue(radioName)).
		AddLabel("Fréquence", discord.NewShortTextInput("frequency").WithRequired(true).WithValue(currentFreq))

	if err := e.Modal(modal); err != nil {
		logger.Error("Error responding with modal", "error", err)
	}
}

func HandleRadioRemoveSelect(e *events.ComponentInteractionCreate) {
	customID := e.Data.CustomID()
	parts := strings.SplitN(customID, "--", 3)
	if len(parts) < 3 {
		respondEphemeral(e, "Erreur: identifiant invalide.")
		return
	}

	channelIDStr := parts[1]
	messageIDStr := parts[2]

	selectData := e.StringSelectMenuInteractionData()
	values := selectData.Values
	if len(values) == 0 {
		respondEphemeral(e, "Aucune radio sélectionnée.")
		return
	}
	selectedName := values[0]

	chanID := snowflake.MustParse(channelIDStr)
	msgID := snowflake.MustParse(messageIDStr)

	msg, err := e.Client().Rest.GetMessage(chanID, msgID)
	if err != nil {
		logger.Error("Error fetching message", "error", err)
		respondEphemeral(e, "Erreur lors de la récupération du message.")
		return
	}

	radios := ParseRadiosFromComponents(msg.Components)
	var newRadios []RadioEntry
	for _, r := range radios {
		if r.Name != selectedName {
			newRadios = append(newRadios, r)
		}
	}

	components := BuildRadioComponents(newRadios)
	if _, err := e.Client().Rest.UpdateMessage(chanID, msgID, discord.NewMessageUpdateV2(components...)); err != nil {
		logger.Error("Error editing radio message", "error", err)
		respondEphemeral(e, "Erreur lors de la modification du message.")
		return
	}

	respondEphemeral(e, "Radio supprimée avec succès.")
}
