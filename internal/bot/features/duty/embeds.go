package duty

import (
	"fmt"
	"strings"
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

func BuildDutyEmbed(onDuty, onCall, offRadio []string) (*discordgo.MessageEmbed, discordgo.ActionsRow) {
	embed := BaseEmbed()
	embed.Title = "Gestionnaire de service"
	embed.Color = 0x0099FF

	dutyList := formatList(onDuty, "Personne n'est en service :(")
	onCallList := formatList(onCall, "Personne n'est en semi service :(")
	offRadioList := formatList(offRadio, "Personne n'est off radio")

	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: fmt.Sprintf("En service (%d) :", len(onDuty)), Value: dutyList, Inline: true},
		{Name: fmt.Sprintf("En semi service (%d) :", len(onCall)), Value: onCallList, Inline: true},
		{Name: fmt.Sprintf("Off radio (%d) :", len(offRadio)), Value: offRadioList, Inline: true},
	}
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "⚠️ Cela peut prendre 5 secondes pour que les changements soient pris en compte.",
	}

	row := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Prendre/Quitter le service",
				Style:    discordgo.PrimaryButton,
				CustomID: "handleLsmsDuty",
			},
			discordgo.Button{
				Label:    "Prendre/Quitter le semi service",
				Style:    discordgo.SecondaryButton,
				CustomID: "handleLsmsOnCall",
			},
			discordgo.Button{
				Label:    "Off radio",
				Style:    discordgo.DangerButton,
				CustomID: "handleLsmsOffRadio",
			},
		},
	}

	return embed, row
}

func formatList(ids []string, empty string) string {
	if len(ids) == 0 {
		return empty
	}
	var mentions []string
	for _, id := range ids {
		mentions = append(mentions, fmt.Sprintf("<@%s>", id))
	}
	return strings.Join(mentions, "\n")
}

func BuildDutyUpdateEmbed(userID string, take bool) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	if take {
		embed.Title = "Prise de service"
		embed.Color = 0x00FF00
		embed.Description = fmt.Sprintf("<@%s> a pris le service.", userID)
	} else {
		embed.Title = "Fin de service"
		embed.Color = 0xFF0000
		embed.Description = fmt.Sprintf("<@%s> a quitté le service.", userID)
	}
	return embed
}

func BuildOnCallUpdateEmbed(userID string, take bool) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	if take {
		embed.Title = "Début du semi service"
		embed.Color = 0x00FF00
		embed.Description = fmt.Sprintf("<@%s> a pris le semi service.", userID)
	} else {
		embed.Title = "Fin du semi service"
		embed.Color = 0xFF0000
		embed.Description = fmt.Sprintf("<@%s> a quitté le semi service.", userID)
	}
	return embed
}

func BuildOffRadioUpdateEmbed(userID string, take bool) *discordgo.MessageEmbed {
	embed := BaseEmbed()
	if take {
		embed.Title = "Passage off radio"
		embed.Color = 0xFF8800
		embed.Description = fmt.Sprintf("<@%s> est passé off radio.", userID)
	} else {
		embed.Title = "Fin du off radio"
		embed.Color = 0x00FF00
		embed.Description = fmt.Sprintf("<@%s> est revenu sur la radio.", userID)
	}
	return embed
}
