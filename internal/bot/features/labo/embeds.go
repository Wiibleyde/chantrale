package labo

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func BaseEmbed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Chantrale",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

func GetAnalyseTypeName(t string) string {
	switch t {
	case "bloodgroup":
		return "Groupe Sanguin"
	case "alcohole":
		return "Taux d'Alcoolémie"
	case "drugs":
		return "Drogues"
	case "diseases":
		return "Maladies"
	default:
		return t
	}
}

func BuildLaboWaitingEmbed(entry *LaboEntry) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	embed.Title = "Analyse en cours"
	embed.Color = 0xFF8800
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: true},
		{Name: "Patient", Value: entry.Name, Inline: true},
		{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: true},
		{Name: "Durée estimée", Value: fmt.Sprintf("%d minute(s)", entry.Time), Inline: true},
	}
	return embed
}

func BuildLaboResultEmbed(entry *LaboEntry) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	embed.Title = "Analyse terminée"
	embed.Color = 0x00FF00
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: true},
		{Name: "Résultat", Value: entry.Result, Inline: true},
		{Name: "Patient", Value: entry.Name, Inline: true},
		{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: true},
	}
	return embed
}

func BuildLaboCancelledEmbed(entry *LaboEntry) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	embed.Title = "Analyse annulée"
	embed.Color = 0xFF0000
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: true},
		{Name: "Patient", Value: entry.Name, Inline: true},
		{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: true},
	}
	return embed
}
