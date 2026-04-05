package embeds

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

func Init(s *discordgo.Session) {
	session = s
}

func BaseEmbed() *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Color:     0xFFFFFF,
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Chantrale – Toujours prête à vous aider.",
		},
	}
	if session != nil && session.State != nil && session.State.User != nil {
		embed.Footer.IconURL = session.State.User.AvatarURL("256")
	}
	return embed
}
