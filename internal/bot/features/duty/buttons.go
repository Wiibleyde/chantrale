package duty

import (
"LsmsBot/internal/database"
"LsmsBot/internal/database/models"
"LsmsBot/internal/logger"

"github.com/disgoorg/disgo/discord"
"github.com/disgoorg/disgo/events"
"github.com/disgoorg/snowflake/v2"
)

func HandleDutyButton(e *events.ComponentInteractionCreate) {
handleRoleToggle(e, "duty")
}

func HandleOnCallButton(e *events.ComponentInteractionCreate) {
handleRoleToggle(e, "oncall")
}

func HandleOffRadioButton(e *events.ComponentInteractionCreate) {
handleRoleToggle(e, "offradio")
}

func handleRoleToggle(e *events.ComponentInteractionCreate, roleType string) {
guildID := *e.GuildID()
msgID := e.Message.ID.String()

var dm models.DutyManager
if err := database.DB.Where("guild_id = ? AND message_id = ?", guildID.String(), msgID).First(&dm).Error; err != nil {
respondEphemeral(e, "Gestionnaire de service introuvable.")
return
}

var roleID *string
var oppositeRoleID *string
switch roleType {
case "duty":
roleID = dm.DutyRoleID
oppositeRoleID = dm.OnCallRoleID
case "oncall":
roleID = dm.OnCallRoleID
oppositeRoleID = dm.DutyRoleID
case "offradio":
roleID = dm.OffRadioRoleID
}

if roleID == nil {
respondEphemeral(e, "Rôle non configuré.")
return
}

member := e.Member()
if member == nil {
respondEphemeral(e, "Erreur: membre introuvable.")
return
}
userID := member.User.ID

roleSnowflake, err := snowflake.Parse(*roleID)
if err != nil {
respondEphemeral(e, "Erreur: rôle invalide.")
return
}

hasRole := false
hasOpposite := false
for _, r := range member.RoleIDs {
if r == roleSnowflake {
hasRole = true
}
if oppositeRoleID != nil {
oppSnowflake, err := snowflake.Parse(*oppositeRoleID)
if err == nil && r == oppSnowflake {
hasOpposite = true
}
}
}

client := e.Client()
var msgContent string
if hasRole {
if err := client.Rest.RemoveMemberRole(guildID, userID, roleSnowflake); err != nil {
logger.Error("Error removing role", "error", err)
respondEphemeral(e, "Erreur lors de la modification du rôle.")
return
}
switch roleType {
case "duty":
msgContent = "Vous avez quitté le service."
case "oncall":
msgContent = "Vous avez quitté le semi service."
case "offradio":
msgContent = "Vous avez quitté le off radio."
}
} else {
if hasOpposite && oppositeRoleID != nil {
oppSnowflake, err := snowflake.Parse(*oppositeRoleID)
if err == nil {
if err := client.Rest.RemoveMemberRole(guildID, userID, oppSnowflake); err != nil {
logger.Error("Error removing opposite role", "error", err)
respondEphemeral(e, "Erreur lors de la modification du rôle.")
return
}
}
}
if err := client.Rest.AddMemberRole(guildID, userID, roleSnowflake); err != nil {
logger.Error("Error adding role", "error", err)
respondEphemeral(e, "Erreur lors de la modification du rôle.")
return
}
switch roleType {
case "duty":
msgContent = "Vous avez pris le service."
case "oncall":
msgContent = "Vous avez pris le semi service."
case "offradio":
msgContent = "Vous êtes passé off radio."
}
}

members, err := client.Rest.GetMembers(guildID, 1000, 0)
if err != nil {
logger.Error("Error fetching members", "error", err)
}

var onDuty, onCall, offRadio []string
if dm.DutyRoleID != nil {
if rid, err := snowflake.Parse(*dm.DutyRoleID); err == nil {
onDuty = membersWithRole(members, rid)
}
}
if dm.OnCallRoleID != nil {
if rid, err := snowflake.Parse(*dm.OnCallRoleID); err == nil {
onCall = membersWithRole(members, rid)
}
}
if dm.OffRadioRoleID != nil {
if rid, err := snowflake.Parse(*dm.OffRadioRoleID); err == nil {
offRadio = membersWithRole(members, rid)
}
}

embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)

chanID, err := snowflake.Parse(dm.ChannelID)
if err == nil {
msgSnowflake, err2 := snowflake.Parse(*dm.MessageID)
if err2 == nil {
embeds := []discord.Embed{embed}
components := []discord.LayoutComponent{row}
if _, err3 := client.Rest.UpdateMessage(chanID, msgSnowflake, discord.MessageUpdate{
Embeds:     &embeds,
Components: &components,
}); err3 != nil {
logger.Error("Error editing duty message", "error", err3)
}
}
}

respondEphemeral(e, msgContent)
}
