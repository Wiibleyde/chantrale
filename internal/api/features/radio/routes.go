package radio

import (
	botradio "LsmsBot/internal/bot/features/radio"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
)

// handleRadios fetches a radio manager message from Discord and returns the
// parsed radio list. The caller must provide the channelID and messageID of
// the message created by the /radio Discord command.
//
// Note: radio state is stored exclusively in the Discord message components
// (no database model). The guildID path param is used for namespace consistency
// but is not validated against the channel — this is an internal API.
func handleRadios(client *bot.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		channelIDStr := c.Params("channelID")
		channelID, err := snowflake.Parse(channelIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "ChannelID invalide.")
		}

		messageIDStr := c.Params("messageID")
		messageID, err := snowflake.Parse(messageIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "MessageID invalide.")
		}

		msg, err := client.Rest.GetMessage(channelID, messageID)
		if err != nil {
			logger.Error("Error fetching radio message", "channelID", channelIDStr, "messageID", messageIDStr, "error", err)
			return fiber.NewError(fiber.StatusNotFound, "Message introuvable.")
		}

		radios := botradio.ParseRadiosFromComponents(msg.Components)

		return c.JSON(fiber.Map{
			"channel_id": channelIDStr,
			"message_id": messageIDStr,
			"radios":     radios,
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
