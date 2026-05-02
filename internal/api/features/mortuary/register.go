package mortuary

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
)

// Register mounts the mortuary read-only routes onto the given router group.
//
//	GET /guilds/:guildID/mortuary/assignments — occupied locker list
//	GET /guilds/:guildID/mortuary/free        — available locker numbers
func Register(r fiber.Router, client *bot.Client) {
	g := r.Group("/guilds/:guildID/mortuary")
	g.Get("/assignments", handleAssignments(client))
	g.Get("/free", handleFree(client))
}
