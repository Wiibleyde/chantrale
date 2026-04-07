package duty

import (
"fmt"
"strings"
"time"

"LsmsBot/internal/bot/embeds"

"github.com/disgoorg/disgo/discord"
"github.com/disgoorg/omit"
)

func BuildDutyEmbed(onDuty, onCall, offRadio []string) (discord.Embed, discord.ActionRowComponent) {
embed := embeds.BaseEmbed()
embed.Title = "Gestionnaire de service"
embed.Color = 0x0099FF

dutyList := formatList(onDuty, "Personne n'est en service :(")
onCallList := formatList(onCall, "Personne n'est en semi service :(")
offRadioList := formatList(offRadio, "Personne n'est off radio")

embed.Fields = []discord.EmbedField{
{Name: fmt.Sprintf("En service (%d) :", len(onDuty)), Value: dutyList, Inline: omit.Ptr(true)},
{Name: fmt.Sprintf("En semi service (%d) :", len(onCall)), Value: onCallList, Inline: omit.Ptr(true)},
{Name: fmt.Sprintf("Off radio (%d) :", len(offRadio)), Value: offRadioList, Inline: omit.Ptr(true)},
}

row := discord.ActionRowComponent{
Components: []discord.InteractiveComponent{
discord.ButtonComponent{Label: "Prendre/Quitter le service", Style: discord.ButtonStylePrimary, CustomID: "handleLsmsDuty"},
discord.ButtonComponent{Label: "Prendre/Quitter le semi service", Style: discord.ButtonStyleSecondary, CustomID: "handleLsmsOnCall"},
discord.ButtonComponent{Label: "Off radio", Style: discord.ButtonStyleDanger, CustomID: "handleLsmsOffRadio"},
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

func BuildDutyUpdateEmbed(userID string, take bool) discord.Embed {
embed := embeds.BaseEmbed()
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

func BuildOnCallUpdateEmbed(userID string, take bool) discord.Embed {
embed := embeds.BaseEmbed()
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

func BuildSummaryEmbed(from, to time.Time, onDuty, onCall, offRadio []string) discord.Embed {
embed := embeds.BaseEmbed()
embed.Title = "Récapitulatif du service"
embed.Color = 0x5865F2
embed.Description = fmt.Sprintf("Période du <t:%d:f> au <t:%d:f>", from.Unix(), to.Unix())
embed.Fields = []discord.EmbedField{
{Name: "Service", Value: formatList(onDuty, "Aucun :("), Inline: omit.Ptr(false)},
{Name: "Semi service", Value: formatList(onCall, "Aucun :("), Inline: omit.Ptr(false)},
{Name: "Off radio", Value: formatList(offRadio, "Aucun"), Inline: omit.Ptr(false)},
}
return embed
}

func BuildOffRadioUpdateEmbed(userID string, take bool) discord.Embed {
embed := embeds.BaseEmbed()
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
