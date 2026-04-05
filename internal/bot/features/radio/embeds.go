package radio

import (
	"encoding/base64"

	"LsmsBot/internal/bot/embeds"

	"github.com/bwmarrin/discordgo"
)

type RadioEntry struct {
	Name      string
	Frequency string
}

func BuildRadioEmbed(radios []RadioEntry) *discordgo.MessageEmbed {
	embed := embeds.BaseEmbed()
	embed.Title = "Gestionnaire de radios"
	embed.Color = 0x0099FF

	if len(radios) == 0 {
		embed.Description = "Aucune radio configurée pour le moment."
	} else {
		embed.Description = "Utilisez les boutons ci-dessous pour gérer les radios disponibles."
		for _, r := range radios {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   r.Name,
				Value:  r.Frequency,
				Inline: true,
			})
		}
	}

	return embed
}

func BuildRadioComponents(radios []RadioEntry) []discordgo.MessageComponent {
	var components []discordgo.MessageComponent

	maxRadios := 20
	if len(radios) < maxRadios {
		maxRadios = len(radios)
	}

	for i := 0; i < maxRadios; i += 5 {
		end := i + 5
		if end > maxRadios {
			end = maxRadios
		}
		row := discordgo.ActionsRow{}
		for _, r := range radios[i:end] {
			encoded := base64.RawURLEncoding.EncodeToString([]byte(r.Name))
			row.Components = append(row.Components, discordgo.Button{
				Label:    "Modifier la radio " + r.Name,
				Style:    discordgo.SecondaryButton,
				CustomID: "lsmsRadioEdit--" + encoded,
			})
		}
		components = append(components, row)
	}

	lastRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "+",
				Style:    discordgo.SuccessButton,
				CustomID: "lsmsRadioAdd",
			},
			discordgo.Button{
				Label:    "-",
				Style:    discordgo.DangerButton,
				CustomID: "lsmsRadioRemove",
			},
		},
	}
	components = append(components, lastRow)

	return components
}

func ParseRadiosFromEmbed(embeds []*discordgo.MessageEmbed) []RadioEntry {
	var radios []RadioEntry
	if len(embeds) == 0 {
		return radios
	}
	for _, field := range embeds[0].Fields {
		radios = append(radios, RadioEntry{
			Name:      field.Name,
			Frequency: field.Value,
		})
	}
	return radios
}
