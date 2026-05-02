package duty

import (
	"errors"

	apihelpers "LsmsBot/internal/api/helpers"
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// handleDutyRole returns a handler that resolves the DutyManager for the
// requested guild and returns members assigned to the role selected by getRoleID.
func handleDutyRole(client *bot.Client, getRoleID func(dm models.DutyManager) *string) fiber.Handler {
	return func(c fiber.Ctx) error {
		guildID, err := parseGuildID(c)
		if err != nil {
			return err
		}

		var dm models.DutyManager
		if dbErr := database.DB.Where("guild_id = ?", guildID.String()).First(&dm).Error; dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "Gestionnaire de service introuvable pour ce serveur.")
			}
			logger.Error("Error fetching duty manager", "error", dbErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur interne du serveur.")
		}

		roleIDPtr := getRoleID(dm)
		if roleIDPtr == nil {
			return fiber.NewError(fiber.StatusNotFound, "Rôle non configuré pour ce gestionnaire.")
		}

		roleID, parseErr := snowflake.Parse(*roleIDPtr)
		if parseErr != nil {
			logger.Error("Error parsing role ID", "error", parseErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur interne du serveur.")
		}

		members, fetchErr := apihelpers.MembersWithRole(client, guildID, roleID)
		if fetchErr != nil {
			logger.Error("Error fetching members", "error", fetchErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur lors de la récupération des membres.")
		}

		return c.JSON(fiber.Map{
			"guild_id": guildID.String(),
			"members":  members,
		})
	}
}

func parseGuildID(c fiber.Ctx) (snowflake.ID, error) {
	guildID, err := snowflake.Parse(c.Params("guildID"))
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "GuildID invalide.")
	}
	return guildID, nil
}
