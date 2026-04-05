package doctor

import (
	"LsmsBot/internal/bot/embeds"

	"github.com/bwmarrin/discordgo"
)

func BuildDossierEmbed(displayName string) *discordgo.MessageEmbed {
	embed := embeds.BaseEmbed()
	embed.Title = "Dossier de formation de " + displayName
	return embed
}

func BuildFormationEmbed(title string) *discordgo.MessageEmbed {
	embed := embeds.BaseEmbed()
	embed.Title = title
	return embed
}
