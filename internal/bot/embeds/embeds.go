package embeds

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
)

var avatarURL string

func Init(url string) {
	avatarURL = url
}

func BaseEmbed() discord.Embed {
	t := time.Now()
	return discord.Embed{
		Color:     0xFFFFFF,
		Timestamp: &t,
		Footer: &discord.EmbedFooter{
			Text:    "Chantrale – LSMS",
			IconURL: avatarURL,
		},
	}
}

func NewContainerV2(accentColor int, components ...discord.ContainerSubComponent) discord.ContainerComponent {
	footer := discord.NewTextDisplay(fmt.Sprintf("-# Chantrale – LSMS  •  <t:%d:f>", time.Now().Unix()))
	components = append(components, discord.NewSmallSeparator(), footer)
	return discord.NewContainer(components...).WithAccentColor(accentColor)
}
