package bed

import (
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

func HandleRemoveBed(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bedLetter := bedLetterFromCustomID(i.MessageComponentData().CustomID)
	if bedLetter == "" {
		respondEphemeral(s, i, "Identifiant de lit invalide.")
		return
	}

	var bm models.BedManager
	if err := database.DB.Where("guild_id = ?", i.GuildID).First(&bm).Error; err != nil {
		respondEphemeral(s, i, "Aucun panneau des lits trouvé.")
		return
	}

	result := database.DB.Where("guild_id = ? AND bed_letter = ?", i.GuildID, bedLetter).Delete(&models.BedAssignment{})
	if result.Error != nil {
		logger.Error("Error deleting bed assignment", "error", result.Error)
		respondEphemeral(s, i, "Erreur lors de la suppression du patient.")
		return
	}
	if result.RowsAffected == 0 {
		respondEphemeral(s, i, "Ce lit est déjà vide.")
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		logger.Error("Error deferring interaction", "error", err)
		return
	}

	if err := updateBedPanel(s, bm); err != nil {
		logger.Error("Error updating bed panel after removal", "error", err)
	}
}
