package duty

import (
"LsmsBot/internal/database"
"LsmsBot/internal/database/models"
"LsmsBot/internal/logger"

"github.com/disgoorg/disgo/discord"
"github.com/disgoorg/disgo/events"
"github.com/disgoorg/disgo/rest"
"github.com/disgoorg/snowflake/v2"
)

var Commands = []discord.ApplicationCommandCreate{
discord.SlashCommandCreate{
Name:        "duty",
Description: "Gérer le gestionnaire de service",
Options: []discord.ApplicationCommandOption{
discord.ApplicationCommandOptionSubCommand{
Name:        "add",
Description: "Ajouter un gestionnaire de service",
Options: []discord.ApplicationCommandOption{
discord.ApplicationCommandOptionRole{Name: "duty", Description: "Rôle de service", Required: true},
discord.ApplicationCommandOptionRole{Name: "oncall", Description: "Rôle de semi service", Required: true},
discord.ApplicationCommandOptionRole{Name: "offradio", Description: "Rôle off radio", Required: true},
discord.ApplicationCommandOptionChannel{Name: "logchannel", Description: "Canal de logs", Required: true},
},
},
discord.ApplicationCommandOptionSubCommand{
Name:        "remove",
Description: "Supprimer un gestionnaire de service",
Options: []discord.ApplicationCommandOption{
discord.ApplicationCommandOptionString{Name: "messageid", Description: "ID du message du gestionnaire", Required: true},
},
},
},
},
}

func respondEphemeral(r interface {
CreateMessage(discord.MessageCreate, ...rest.RequestOpt) error
}, content string) {
if err := r.CreateMessage(discord.MessageCreate{
Content: content,
Flags:   discord.MessageFlagEphemeral,
}); err != nil {
logger.Error("Error responding to interaction", "error", err)
}
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
data := e.SlashCommandInteractionData()
if data.SubCommandName == nil {
return
}
switch *data.SubCommandName {
case "add":
handleAdd(e)
case "remove":
handleRemove(e)
}
}

func handleAdd(e *events.ApplicationCommandInteractionCreate) {
member := e.Member()
if member == nil || !member.Permissions.Has(discord.PermissionManageChannels) {
respondEphemeral(e, "Vous n'avez pas la permission de gérer les canaux.")
return
}

data := e.SlashCommandInteractionData()
dutyRole, _ := data.OptRole("duty")
onCallRole, _ := data.OptRole("oncall")
offRadioRole, _ := data.OptRole("offradio")
logChannel, _ := data.OptChannel("logchannel")

client := e.Client()
guildID := *e.GuildID()

botMember, err := client.Rest.GetMember(guildID, client.ID())
if err != nil {
respondEphemeral(e, "Erreur lors de la récupération des informations du bot.")
return
}

guild, err := client.Rest.GetGuild(guildID, false)
if err != nil {
respondEphemeral(e, "Erreur lors de la récupération du serveur.")
return
}

botHighest := highestRolePosition(guild.Roles, botMember.RoleIDs)

for _, r := range []discord.Role{dutyRole, onCallRole, offRadioRole} {
if r.Position >= botHighest {
respondEphemeral(e, "Un ou plusieurs rôles sont au-dessus ou au même niveau que le rôle le plus haut du bot.")
return
}
}

members, err := client.Rest.GetMembers(guildID, 1000, 0)
if err != nil {
respondEphemeral(e, "Erreur lors de la récupération des membres.")
return
}

onDuty := membersWithRole(members, dutyRole.ID)
onCall := membersWithRole(members, onCallRole.ID)
offRadio := membersWithRole(members, offRadioRole.ID)

embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)

channelID := e.Channel().ID()
msg, err := client.Rest.CreateMessage(channelID, discord.MessageCreate{
Embeds:     []discord.Embed{embed},
Components: []discord.LayoutComponent{row},
})
if err != nil {
logger.Error("Error sending duty message", "error", err)
respondEphemeral(e, "Erreur lors de l'envoi du message.")
return
}

dutyRoleID := dutyRole.ID.String()
onCallRoleID := onCallRole.ID.String()
offRadioRoleID := offRadioRole.ID.String()
logsChannelID := logChannel.ID.String()
msgID := msg.ID.String()

dm := models.DutyManager{
GuildID:        guildID.String(),
ChannelID:      channelID.String(),
MessageID:      &msgID,
DutyRoleID:     &dutyRoleID,
OnCallRoleID:   &onCallRoleID,
OffRadioRoleID: &offRadioRoleID,
LogsChannelID:  &logsChannelID,
}

if err := database.DB.Create(&dm).Error; err != nil {
logger.Error("Error saving DutyManager", "error", err)
respondEphemeral(e, "Erreur lors de l'enregistrement en base de données.")
return
}

respondEphemeral(e, "Gestionnaire de service créé avec succès.")
}

func handleRemove(e *events.ApplicationCommandInteractionCreate) {
member := e.Member()
if member == nil || !member.Permissions.Has(discord.PermissionManageChannels) {
respondEphemeral(e, "Vous n'avez pas la permission de gérer les canaux.")
return
}

data := e.SlashCommandInteractionData()
messageID := data.String("messageid")

guildID := *e.GuildID()

var dm models.DutyManager
if err := database.DB.Where("guild_id = ? AND message_id = ?", guildID.String(), messageID).First(&dm).Error; err != nil {
respondEphemeral(e, "Gestionnaire de service introuvable.")
return
}

chanID, err := snowflake.Parse(dm.ChannelID)
if err == nil {
msgSnowflake, err2 := snowflake.Parse(messageID)
if err2 == nil {
if err3 := e.Client().Rest.DeleteMessage(chanID, msgSnowflake); err3 != nil {
logger.Error("Error deleting duty message", "error", err3)
}
}
}

if err := database.DB.Delete(&dm).Error; err != nil {
logger.Error("Error deleting DutyManager", "error", err)
respondEphemeral(e, "Erreur lors de la suppression.")
return
}

respondEphemeral(e, "Gestionnaire de service supprimé avec succès.")
}

func membersWithRole(members []discord.Member, roleID snowflake.ID) []string {
var result []string
for _, m := range members {
for _, r := range m.RoleIDs {
if r == roleID {
result = append(result, m.User.ID.String())
break
}
}
}
return result
}

func highestRolePosition(roles []discord.Role, memberRoleIDs []snowflake.ID) int {
highest := 0
for _, r := range roles {
for _, id := range memberRoleIDs {
if r.ID == id && r.Position > highest {
highest = r.Position
}
}
}
return highest
}
