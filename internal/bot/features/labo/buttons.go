package labo

import (
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
)

func HandleCancelButton(e *events.ComponentInteractionCreate) {
	messageID := e.Message.ID

	found, entry := Queue.CancelByMessageID(messageID)
	if !found {
		respondEphemeral(e, "Analyse introuvable ou déjà terminée.")
		return
	}

	member := e.Member()
	if member == nil || entry.UserID != member.User.ID {
		Queue.Add(entry)
		respondEphemeral(e, "Vous ne pouvez pas annuler l'analyse d'un autre membre.")
		return
	}

	cancelComponents := BuildLaboCancelledComponents(entry)
	if _, err := e.Client().Rest.UpdateMessage(e.Channel().ID(), messageID, discord.NewMessageUpdateV2(cancelComponents...)); err != nil {
		logger.Error("Error editing labo message", "error", err)
	}

	respondEphemeral(e, "Analyse annulée avec succès.")
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
