package mortuary

import (
	"fmt"
	"strings"

	"LsmsBot/internal/bot/embeds"
	"LsmsBot/internal/database/models"

	"github.com/disgoorg/disgo/discord"
)

const totalLockers = 12

func BuildMortuaryEmbed(assignments []models.MortuaryAssignment) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "🪦 Morgue — Casiers frigorifiques"
	embed.Color = 0x2F3136

	aMap := make(map[int]models.MortuaryAssignment, len(assignments))
	for _, a := range assignments {
		aMap[a.LockerNumber] = a
	}

	var lines []string
	for i := 1; i <= totalLockers; i++ {
		if a, ok := aMap[i]; ok {
			line := fmt.Sprintf("**Casier %d** — ||%s||", i, a.Name)
			if a.Comment != nil && *a.Comment != "" {
				line += fmt.Sprintf(" · *%s*", *a.Comment)
			}
			lines = append(lines, line)
		} else {
			lines = append(lines, fmt.Sprintf("**Casier %d** — *Vide*", i))
		}
	}

	desc := strings.Join(lines, "\n")
	desc += "\n\nUtilisez `/mortuary add` pour ajouter un corps à un casier."
	embed.Description = desc

	return embed
}
