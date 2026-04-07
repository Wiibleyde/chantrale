package labo

import (
"fmt"

"LsmsBot/internal/bot/embeds"

"github.com/disgoorg/disgo/discord"
"github.com/disgoorg/omit"
)

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

func BuildLaboWaitingEmbed(entry *LaboEntry) discord.Embed {
embed := embeds.BaseEmbed()
embed.Title = "Analyse en cours"
embed.Color = 0xFF8800
embed.Fields = []discord.EmbedField{
{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: omit.Ptr(true)},
{Name: "Patient", Value: entry.Name, Inline: omit.Ptr(true)},
{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: omit.Ptr(true)},
{Name: "Durée estimée", Value: fmt.Sprintf("%d minute(s)", entry.Time), Inline: omit.Ptr(true)},
}
return embed
}

func BuildLaboResultEmbed(entry *LaboEntry) discord.Embed {
embed := embeds.BaseEmbed()
embed.Title = "Analyse terminée"
embed.Color = 0x00FF00
embed.Fields = []discord.EmbedField{
{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: omit.Ptr(true)},
{Name: "Résultat", Value: entry.Result, Inline: omit.Ptr(true)},
{Name: "Patient", Value: entry.Name, Inline: omit.Ptr(true)},
{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: omit.Ptr(true)},
}
return embed
}

func BuildLaboCancelledEmbed(entry *LaboEntry) discord.Embed {
embed := embeds.BaseEmbed()
embed.Title = "Analyse annulée"
embed.Color = 0xFF0000
embed.Fields = []discord.EmbedField{
{Name: "Type", Value: GetAnalyseTypeName(entry.Type), Inline: omit.Ptr(true)},
{Name: "Patient", Value: entry.Name, Inline: omit.Ptr(true)},
{Name: "Demandeur", Value: fmt.Sprintf("<@%s>", entry.UserID), Inline: omit.Ptr(true)},
}
return embed
}
