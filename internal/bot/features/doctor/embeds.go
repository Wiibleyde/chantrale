package doctor

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

func BaseEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Chantrale",
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     0x0099FF,
	}
}

func BuildDossierEmbed(displayName string) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	embed.Title = "Dossier de formation de " + displayName
	return embed
}

func BuildFormationEmbed(title string) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	embed.Title = title
	return embed
}
