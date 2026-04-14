package bed

import (
	"fmt"
	"strings"

	"LsmsBot/internal/bot/embeds"
	"LsmsBot/internal/database/models"

	"github.com/disgoorg/disgo/discord"
)

func BuildBedEmbed(assignments []models.BedAssignment) discord.Embed {
	embed := embeds.BaseEmbed()
	embed.Title = "Gestion des lits"
	embed.Color = 0x0099FF
	embed.Image = &discord.EmbedResource{URL: "attachment://beds.png"}

	var lines []string
	for _, a := range assignments {
		if a.Description != nil && *a.Description != "" {
			lines = append(lines, fmt.Sprintf("**Lit %s** — *%s* : %s", a.BedLetter, a.Name, *a.Description))
		}
	}

	desc := strings.Join(lines, "\n")
	if desc != "" {
		desc += "\n\n"
	}
	desc += "Utilisez `/beds add` pour ajouter un patient à un lit."
	embed.Description = desc

	return embed
}
