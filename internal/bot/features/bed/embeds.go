package bed

import (
	"LsmsBot/internal/bot/embeds"

	"github.com/bwmarrin/discordgo"
)

func BuildBedEmbed() *discordgo.MessageEmbed {
	embed := embeds.BaseEmbed()
	embed.Title = "Gestion des lits"
	embed.Description = "Utilisez `/beds add` pour ajouter un patient à un lit."
	embed.Color = 0x0099FF
	embed.Image = &discordgo.MessageEmbedImage{
		URL: "attachment://beds.png",
	}
	return embed
}
