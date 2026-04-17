package radio

import (
	"LsmsBot/internal/bot/helpers"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "radio",
		Description: "Créer un gestionnaire de radios",
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	if !helpers.RequirePermission(e, discord.PermissionManageChannels, "Vous n'avez pas la permission de gérer les canaux.") {
		return
	}

	components := BuildRadioComponents(nil)

	channelID := e.Channel().ID()
	if _, err := e.Client().Rest.CreateMessage(channelID, discord.NewMessageCreateV2(components...)); err != nil {
		logger.Error("Error sending radio message", "error", err)
		helpers.RespondEphemeral(e, "Erreur lors de l'envoi du message.")
		return
	}

	stats.Record(e.GuildID().String(), e.Member().User.ID.String(), "radio.setup", map[string]any{
		"channel_id": channelID.String(),
	})

	helpers.RespondEphemeral(e, "Gestionnaire de radios créé avec succès.")
}
