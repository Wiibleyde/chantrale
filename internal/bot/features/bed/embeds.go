package bed

import (
	"LsmsBot/internal/bot/embeds"

	"github.com/disgoorg/disgo/discord"
)

func BuildBedEmbed() discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Gestion des lits"
	embed.Description = "Utilisez `/beds add` pour ajouter un patient à un lit."
	embed.Color = 0x0099FF
	embed.Image = &discord.EmbedResource{URL: "attachment://beds.png"}
	return embed
}
