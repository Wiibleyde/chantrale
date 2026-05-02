package bed

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
)

// Register mounts the bed read-only routes onto the given router group.
//
//	GET /guilds/:guildID/bed/assignments — occupied bed list
//	GET /guilds/:guildID/bed/free        — available bed letters
func Register(r fiber.Router, client *bot.Client) {
	g := r.Group("/guilds/:guildID/bed")
	g.Get("/assignments", handleAssignments(client))
	g.Get("/free", handleFree(client))
}
