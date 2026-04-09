package radio

import (
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "radio",
		Description: "Créer un gestionnaire de radios",
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	member := e.Member()
	if member == nil || !member.Permissions.Has(discord.PermissionManageChannels) {
		respondEphemeral(e, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	components := BuildRadioComponents(nil)

	channelID := e.Channel().ID()
	if _, err := e.Client().Rest.CreateMessage(channelID, discord.NewMessageCreateV2(components...)); err != nil {
		logger.Error("Error sending radio message", "error", err)
		respondEphemeral(e, "Erreur lors de l'envoi du message.")
		return
	}

	respondEphemeral(e, "Gestionnaire de radios créé avec succès.")
}

func respondEphemeral(r interface {
	CreateMessage(discord.MessageCreate, ...rest.RequestOpt) error
}, content string) {
	if err := r.CreateMessage(discord.MessageCreate{
		Content: content,
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}
