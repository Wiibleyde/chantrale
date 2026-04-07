package radio

import (
	"encoding/base64"

	"LsmsBot/internal/bot/embeds"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/omit"
)

type RadioEntry struct {
	Name      string
	Frequency string
}

func BuildRadioEmbed(radios []RadioEntry) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Gestionnaire de radios"
	embed.Color = 0x0099FF

	if len(radios) == 0 {
		embed.Description = "Aucune radio configurée pour le moment."
	} else {
		embed.Description = "Utilisez les boutons ci-dessous pour gérer les radios disponibles."
		for _, r := range radios {
			embed.Fields = append(embed.Fields, discord.EmbedField{
				Name:   r.Name,
				Value:  r.Frequency,
				Inline: omit.Ptr(true),
			})
		}
	}

	return embed
}

func BuildRadioComponents(radios []RadioEntry) []discord.LayoutComponent {
	var components []discord.LayoutComponent

	maxRadios := 20
	if len(radios) < maxRadios {
		maxRadios = len(radios)
	}

	for i := 0; i < maxRadios; i += 5 {
		end := i + 5
		if end > maxRadios {
			end = maxRadios
		}
		var btns []discord.InteractiveComponent
		for _, r := range radios[i:end] {
			encoded := base64.RawURLEncoding.EncodeToString([]byte(r.Name))
			btns = append(btns, discord.ButtonComponent{
				Label:    "Modifier la radio " + r.Name,
				Style:    discord.ButtonStyleSecondary,
				CustomID: "lsmsRadioEdit--" + encoded,
			})
		}
		components = append(components, discord.ActionRowComponent{Components: btns})
	}

	lastRow := discord.ActionRowComponent{
		Components: []discord.InteractiveComponent{
			discord.ButtonComponent{Label: "+", Style: discord.ButtonStyleSuccess, CustomID: "lsmsRadioAdd"},
			discord.ButtonComponent{Label: "-", Style: discord.ButtonStyleDanger, CustomID: "lsmsRadioRemove"},
		},
	}
	components = append(components, lastRow)

	return components
}

func ParseRadiosFromEmbed(embeds []discord.Embed) []RadioEntry {
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
