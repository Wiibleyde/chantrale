package radio

import (
	"encoding/base64"
	"strings"

	"LsmsBot/internal/bot/embeds"

	"github.com/disgoorg/disgo/discord"
)

type RadioEntry struct {
	Name      string
	Frequency string
}

func BuildRadioComponents(radios []RadioEntry) []discord.LayoutComponent {
	var subs []discord.ContainerSubComponent

	subs = append(subs, discord.NewTextDisplay("## 📻 Gestionnaire de radios"))

	if len(radios) == 0 {
		subs = append(subs,
			discord.NewLargeSeparator(),
			discord.NewTextDisplay("Aucune radio configurée pour le moment."),
		)
	} else {
		for _, r := range radios {
			encoded := base64.RawURLEncoding.EncodeToString([]byte(r.Name))
			section := discord.NewSection(
				discord.NewTextDisplay("**"+r.Name+"**"),
				discord.NewTextDisplay(r.Frequency),
			).WithAccessory(discord.NewSecondaryButton("✏️ Modifier", "lsmsRadioEdit--"+encoded))
			subs = append(subs, discord.NewSmallSeparator(), section)
		}
	}

	subs = append(subs,
		discord.NewLargeSeparator(),
		discord.NewActionRow(
			discord.NewSuccessButton("➕ Ajouter une radio", "lsmsRadioAdd"),
			discord.NewDangerButton("➖ Supprimer une radio", "lsmsRadioRemove"),
		),
	)

	return []discord.LayoutComponent{embeds.NewContainerV2(0x0099FF, subs...)}
}

func ParseRadiosFromComponents(components []discord.LayoutComponent) []RadioEntry {
	var radios []RadioEntry
	for _, layout := range components {
		container, ok := layout.(discord.ContainerComponent)
		if !ok {
			continue
		}
		for _, sub := range container.Components {
			section, ok := sub.(discord.SectionComponent)
			if !ok {
				continue
			}
			btn, ok := section.Accessory.(discord.ButtonComponent)
			if !ok {
				continue
			}
			cid := btn.CustomID
			if !strings.HasPrefix(string(cid), "lsmsRadioEdit--") {
				continue
			}
			encoded := strings.TrimPrefix(string(cid), "lsmsRadioEdit--")
			nameBytes, err := base64.RawURLEncoding.DecodeString(encoded)
			if err != nil {
				continue
			}
			freq := ""
			if len(section.Components) >= 2 {
				if td, ok := section.Components[1].(discord.TextDisplayComponent); ok {
					freq = td.Content
				}
			}
			radios = append(radios, RadioEntry{
				Name:      string(nameBytes),
				Frequency: freq,
			})
		}
	}
	return radios
}
