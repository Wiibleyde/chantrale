package doctor

import (
	"LsmsBot/internal/bot/embeds"

	"github.com/disgoorg/disgo/discord"
)

func BuildDossierEmbed(displayName string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Dossier de formation de " + displayName
	return embed
}

func BuildFormationEmbed(title string) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = title
	return embed
}
