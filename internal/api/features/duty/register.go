package duty

import (
	"LsmsBot/internal/database/models"

	"github.com/disgoorg/disgo/bot"
	"github.com/gofiber/fiber/v3"
)

// Register mounts the duty read-only routes onto the given router group.
//
//	GET /guilds/:guildID/duty/onduty   — members currently on duty
//	GET /guilds/:guildID/duty/oncall   — members on call (astreinte)
//	GET /guilds/:guildID/duty/offradio — members off radio
func Register(r fiber.Router, client *bot.Client) {
	g := r.Group("/guilds/:guildID/duty")
	g.Get("/onduty", handleDutyRole(client, func(dm models.DutyManager) *string { return dm.DutyRoleID }))
	g.Get("/oncall", handleDutyRole(client, func(dm models.DutyManager) *string { return dm.OnCallRoleID }))
	g.Get("/offradio", handleDutyRole(client, func(dm models.DutyManager) *string { return dm.OffRadioRoleID }))
}
