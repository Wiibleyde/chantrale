package bed

import (
	"errors"
	"sort"

	botbed "LsmsBot/internal/bot/features/bed"
	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func handleAssignments(_ *bot.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		guildID, err := parseGuildID(c)
		if err != nil {
			return err
		}

		var assignments []models.BedAssignment
		if dbErr := database.DB.Where("guild_id = ?", guildID.String()).Find(&assignments).Error; dbErr != nil {
			logger.Error("Error fetching bed assignments", "error", dbErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur interne du serveur.")
		}

		return c.JSON(fiber.Map{
			"guild_id":    guildID.String(),
			"assignments": assignments,
		})
	}
}

func handleFree(client *bot.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		guildID, err := parseGuildID(c)
		if err != nil {
			return err
		}

		if dbErr := database.DB.Where("guild_id = ?", guildID.String()).First(&models.BedManager{}).Error; dbErr != nil {
			if errors.Is(dbErr, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "Gestionnaire de lits introuvable pour ce serveur.")
			}
			logger.Error("Error fetching bed manager", "error", dbErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur interne du serveur.")
		}

		var assignments []models.BedAssignment
		if dbErr := database.DB.Where("guild_id = ?", guildID.String()).Find(&assignments).Error; dbErr != nil {
			logger.Error("Error fetching bed assignments", "error", dbErr)
			return fiber.NewError(fiber.StatusInternalServerError, "Erreur interne du serveur.")
		}

		occupied := make(map[string]bool, len(assignments))
		for _, a := range assignments {
			occupied[a.BedLetter] = true
		}

		free := []string{}
		for _, cfg := range botbed.BedConfigs {
			if !occupied[cfg.Letter] {
				free = append(free, cfg.Letter)
			}
		}
		sort.Strings(free)

		return c.JSON(fiber.Map{
			"guild_id": guildID.String(),
			"free":     free,
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
