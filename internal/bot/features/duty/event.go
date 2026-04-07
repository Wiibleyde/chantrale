package duty

import (
	"sync"
	"time"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

var (
	debounceMap = make(map[string]*time.Timer)
	debounceMu  sync.Mutex
)

func HandleGuildMemberUpdate(e *events.GuildMemberUpdate) {
	key := e.GuildID.String() + "-" + e.Member.User.ID.String()

	debounceMu.Lock()
	if t, ok := debounceMap[key]; ok {
		t.Stop()
	}

	prevRoles := e.OldMember.RoleIDs

	client := e.Client()
	guildID := e.GuildID
	member := e.Member

	debounceMap[key] = time.AfterFunc(100*time.Millisecond, func() {
		debounceMu.Lock()
		delete(debounceMap, key)
		debounceMu.Unlock()

		handleMemberRoleChange(client, guildID, member, prevRoles)
	})
	debounceMu.Unlock()
}

func handleMemberRoleChange(client *bot.Client, guildID snowflake.ID, member discord.Member, prevRoles []snowflake.ID) {
	newRoles := member.RoleIDs
	added := diffRoles(prevRoles, newRoles)
	removed := diffRoles(newRoles, prevRoles)
	changed := append(added, removed...)

	if len(changed) == 0 {
		return
	}

	var dms []models.DutyManager
	if err := database.DB.Where("guild_id = ?", guildID.String()).Find(&dms).Error; err != nil {
		logger.Error("Error fetching DutyManagers", "error", err)
		return
	}

	members, err := client.Rest.GetMembers(guildID, 1000, 0)
	if err != nil {
		logger.Error("Error fetching members", "error", err)
		return
	}

	for _, dm := range dms {
		affected := false
		for _, roleID := range changed {
			roleIDStr := roleID.String()
			if (dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr) ||
				(dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr) ||
				(dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr) {
				affected = true
				break
			}
		}
		if !affected {
			continue
		}

		if dm.MessageID == nil {
			continue
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

		if dm.LogsChannelID == nil {
			continue
		}

		logsChannelID, err := snowflake.Parse(*dm.LogsChannelID)
		if err != nil {
			continue
		}

		userID := member.User.ID.String()
		for _, roleID := range added {
			roleIDStr := roleID.String()
			var logEmbed *discord.Embed
			if dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr {
				e := BuildDutyUpdateEmbed(userID, true)
				logEmbed = &e
				trackDuty(guildID.String(), userID)
			} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr {
				e := BuildOnCallUpdateEmbed(userID, true)
				logEmbed = &e
				trackOnCall(guildID.String(), userID)
			} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr {
				e := BuildOffRadioUpdateEmbed(userID, true)
				logEmbed = &e
				trackOffRadio(guildID.String(), userID)
			}
			if logEmbed != nil {
				if _, err := client.Rest.CreateMessage(logsChannelID, discord.MessageCreate{
					Embeds: []discord.Embed{*logEmbed},
				}); err != nil {
					logger.Error("Error sending log embed", "error", err)
				}
			}
		}
		for _, roleID := range removed {
			roleIDStr := roleID.String()
			var logEmbed *discord.Embed
			if dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr {
				e := BuildDutyUpdateEmbed(userID, false)
				logEmbed = &e
			} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr {
				e := BuildOnCallUpdateEmbed(userID, false)
				logEmbed = &e
			} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr {
				e := BuildOffRadioUpdateEmbed(userID, false)
				logEmbed = &e
			}
			if logEmbed != nil {
				if _, err := client.Rest.CreateMessage(logsChannelID, discord.MessageCreate{
					Embeds: []discord.Embed{*logEmbed},
				}); err != nil {
					logger.Error("Error sending log embed", "error", err)
				}
			}
		}
	}
}

func diffRoles(a, b []snowflake.ID) []snowflake.ID {
	set := make(map[snowflake.ID]bool)
	for _, r := range a {
		set[r] = true
	}
	var diff []snowflake.ID
	for _, r := range b {
		if !set[r] {
			diff = append(diff, r)
		}
	}
	return diff
}
