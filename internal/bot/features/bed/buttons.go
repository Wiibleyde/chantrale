package bed

import (
	"LsmsBot/internal/bot/helpers"
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/events"
)

func HandleRemoveBed(e *events.ComponentInteractionCreate) {
	bedLetter := helpers.SuffixFromCustomID(e.Data.CustomID())
	if bedLetter == "" {
		helpers.RespondEphemeral(e, "Identifiant de lit invalide.")
		return
	}

	guildID := *e.GuildID()

	var bm models.BedManager
	if err := database.DB.Where("guild_id = ?", guildID.String()).First(&bm).Error; err != nil {
		helpers.RespondEphemeral(e, "Aucun panneau des lits trouvé.")
		return
	}

	var assignment models.BedAssignment
	if err := database.DB.Where("guild_id = ? AND bed_letter = ?", guildID.String(), bedLetter).First(&assignment).Error; err != nil {
		helpers.RespondEphemeral(e, "Ce lit est déjà vide.")
		return
	}

	result := database.DB.Delete(&assignment)
	if result.Error != nil {
		logger.Error("Error deleting bed assignment", "error", result.Error)
		helpers.RespondEphemeral(e, "Erreur lors de la suppression du patient.")
		return
	}

	userID := ""
	if m := e.Member(); m != nil {
		userID = m.User.ID.String()
	}
	stats.Record(guildID.String(), userID, "bed.release", map[string]any{
		"bed_letter":   bedLetter,
		"patient_name": assignment.Name,
		"under_arrest": assignment.UnderArrest,
		"death":        assignment.Death,
	})

	if err := e.DeferUpdateMessage(); err != nil {
		logger.Error("Error deferring interaction", "error", err)
		return
	}

	if err := updateBedPanel(e.Client(), bm); err != nil {
		logger.Error("Error updating bed panel after removal", "error", err)
	}
}
