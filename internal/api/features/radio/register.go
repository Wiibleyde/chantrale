package radio

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
)

// Register mounts the radio read-only routes onto the given router group.
//
//	GET /guilds/:guildID/radio/:channelID/:messageID — radio list from a manager message
//
// Radio state is stored in Discord message components, not in the database.
// The channelID and messageID correspond to the message created by /radio.
func Register(r fiber.Router, client *bot.Client) {
	r.Get("/guilds/:guildID/radio/:channelID/:messageID", handleRadios(client))
}
