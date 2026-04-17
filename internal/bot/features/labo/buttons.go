package labo

import (
	"time"

	"LsmsBot/internal/bot/helpers"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func HandleCancelButton(e *events.ComponentInteractionCreate) {
	messageID := e.Message.ID

	found, entry := Queue.CancelByMessageID(messageID)
	if !found {
		helpers.RespondEphemeral(e, "Analyse introuvable ou déjà terminée.")
		return
	}

	member := e.Member()
	if member == nil || entry.UserID != member.User.ID {
		Queue.Add(entry)
		helpers.RespondEphemeral(e, "Vous ne pouvez pas annuler l'analyse d'un autre membre.")
		return
	}

	cancelComponents := BuildLaboCancelledComponents(entry)
	if _, err := e.Client().Rest.UpdateMessage(e.Channel().ID(), messageID, discord.NewMessageUpdateV2(cancelComponents...)); err != nil {
		logger.Error("Error editing labo message", "error", err)
	}

	stats.Record(entry.GuildID, member.User.ID.String(), "labo.test_cancel", map[string]any{
		"test_type":       entry.Type,
		"patient_name":    entry.Name,
		"elapsed_seconds": int(time.Since(entry.StartTime).Seconds()),
	})

	helpers.RespondEphemeral(e, "Analyse annulée avec succès.")
}
