package bed

import (
"LsmsBot/internal/database"
"LsmsBot/internal/database/models"
"LsmsBot/internal/logger"

"github.com/disgoorg/disgo/events"
)

func HandleRemoveBed(e *events.ComponentInteractionCreate) {
bedLetter := bedLetterFromCustomID(e.Data.CustomID())
if bedLetter == "" {
respondEphemeral(e, "Identifiant de lit invalide.")
return
}

guildID := *e.GuildID()

var bm models.BedManager
if err := database.DB.Where("guild_id = ?", guildID.String()).First(&bm).Error; err != nil {
respondEphemeral(e, "Aucun panneau des lits trouvé.")
return
}

result := database.DB.Where("guild_id = ? AND bed_letter = ?", guildID.String(), bedLetter).Delete(&models.BedAssignment{})
if result.Error != nil {
logger.Error("Error deleting bed assignment", "error", result.Error)
respondEphemeral(e, "Erreur lors de la suppression du patient.")
return
}
if result.RowsAffected == 0 {
respondEphemeral(e, "Ce lit est déjà vide.")
return
}

if err := e.DeferUpdateMessage(); err != nil {
logger.Error("Error deferring interaction", "error", err)
return
}

if err := updateBedPanel(e.Client(), bm); err != nil {
logger.Error("Error updating bed panel after removal", "error", err)
}
}
