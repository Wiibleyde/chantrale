package mortuary

import (
	"strconv"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/events"
)

func HandleRemoveLocker(e *events.ComponentInteractionCreate) {
	lockerStr := lockerFromCustomID(e.Data.CustomID())
	if lockerStr == "" {
		respondEphemeral(e, "Identifiant de casier invalide.")
		return
	}

	lockerNumber, err := strconv.Atoi(lockerStr)
	if err != nil {
		respondEphemeral(e, "Identifiant de casier invalide.")
		return
	}

	guildID := *e.GuildID()

	var mm models.MortuaryManager
	if err := database.DB.Where("guild_id = ?", guildID.String()).First(&mm).Error; err != nil {
		respondEphemeral(e, "Aucun panneau de la morgue trouvé.")
		return
	}

	var assignment models.MortuaryAssignment
	if err := database.DB.Where("guild_id = ? AND locker_number = ?", guildID.String(), lockerNumber).First(&assignment).Error; err != nil {
		respondEphemeral(e, "Ce casier est déjà vide.")
		return
	}

	result := database.DB.Delete(&assignment)
	if result.Error != nil {
		logger.Error("Error deleting mortuary assignment", "error", result.Error)
		respondEphemeral(e, "Erreur lors de la suppression du corps.")
		return
	}

	userID := ""
	if m := e.Member(); m != nil {
		userID = m.User.ID.String()
	}
	stats.Record(guildID.String(), userID, "mortuary.release", map[string]any{
		"locker_number": lockerNumber,
		"name":          assignment.Name,
	})

	if err := e.DeferUpdateMessage(); err != nil {
		logger.Error("Error deferring interaction", "error", err)
		return
	}

	if err := updateMortuaryPanel(e.Client(), mm); err != nil {
		logger.Error("Error updating mortuary panel after removal", "error", err)
	}
}
