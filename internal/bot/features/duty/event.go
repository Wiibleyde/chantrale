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

// debounceState keeps prevRoles from the first event so the full diff is captured
// even when multiple role changes arrive in rapid succession.
type debounceState struct {
	timer     *time.Timer
	prevRoles []snowflake.ID
}

var (
	debounceMap = make(map[string]*debounceState)
	debounceMu  sync.Mutex

	// embedDebounceMap groups rapid role changes into a single embed update per message.
	embedDebounceMap = make(map[string]*time.Timer)
	embedDebounceMu  sync.Mutex
)

func HandleGuildMemberUpdate(e *events.GuildMemberUpdate) {
	key := e.GuildID.String() + "-" + e.Member.User.ID.String()

	debounceMu.Lock()

	prevRoles := e.OldMember.RoleIDs
	if existing, ok := debounceMap[key]; ok {
		existing.timer.Stop()
		prevRoles = existing.prevRoles
	}

	client := e.Client()
	guildID := e.GuildID
	member := e.Member

	state := &debounceState{prevRoles: prevRoles}
	state.timer = time.AfterFunc(100*time.Millisecond, func() {
		debounceMu.Lock()
		delete(debounceMap, key)
		debounceMu.Unlock()

		handleMemberRoleChange(client, guildID, member, prevRoles)
	})
	debounceMap[key] = state
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
		if !affected || dm.MessageID == nil {
			continue
		}

		if dm.LogsChannelID != nil {
			if logsChannelID, err := snowflake.Parse(*dm.LogsChannelID); err == nil {
				sendLogMessages(client, logsChannelID, guildID.String(), member.User.ID.String(), dm, added, removed)
			}
		}

		scheduleEmbedUpdate(client, guildID, dm)
	}
}

func sendLogMessages(
	client *bot.Client,
	logsChannelID snowflake.ID,
	guildID, userID string,
	dm models.DutyManager,
	added, removed []snowflake.ID,
) {
	send := func(e discord.Embed) {
		if _, err := client.Rest.CreateMessage(logsChannelID, discord.MessageCreate{
			Embeds: []discord.Embed{e},
		}); err != nil {
			logger.Error("Error sending log embed", "error", err)
		}
	}

	for _, roleID := range added {
		roleIDStr := roleID.String()
		if dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr {
			send(BuildDutyUpdateEmbed(userID, true))
			trackDuty(guildID, userID)
		} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr {
			send(BuildOnCallUpdateEmbed(userID, true))
			trackOnCall(guildID, userID)
		} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr {
			send(BuildOffRadioUpdateEmbed(userID, true))
			trackOffRadio(guildID, userID)
		}
	}
	for _, roleID := range removed {
		roleIDStr := roleID.String()
		if dm.DutyRoleID != nil && *dm.DutyRoleID == roleIDStr {
			send(BuildDutyUpdateEmbed(userID, false))
		} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleIDStr {
			send(BuildOnCallUpdateEmbed(userID, false))
		} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleIDStr {
			send(BuildOffRadioUpdateEmbed(userID, false))
		}
	}
}

// scheduleEmbedUpdate batches embed updates within a 500ms window to avoid spamming the API.
func scheduleEmbedUpdate(client *bot.Client, guildID snowflake.ID, dm models.DutyManager) {
	key := *dm.MessageID

	embedDebounceMu.Lock()
	if t, ok := embedDebounceMap[key]; ok {
		t.Stop()
	}
	embedDebounceMap[key] = time.AfterFunc(500*time.Millisecond, func() {
		embedDebounceMu.Lock()
		delete(embedDebounceMap, key)
		embedDebounceMu.Unlock()

		performEmbedUpdate(client, guildID, dm)
	})
	embedDebounceMu.Unlock()
}

// performEmbedUpdate fetches members at call time so the embed is never stale.
func performEmbedUpdate(client *bot.Client, guildID snowflake.ID, dm models.DutyManager) {
	members, err := client.Rest.GetMembers(guildID, 1000, 0)
	if err != nil {
		logger.Error("Error fetching members for embed update", "guild", guildID.String(), "error", err)
		return
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

	setGuildCounts(guildID.String(), len(onDuty), len(offRadio))

	embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)
	chanID, err := snowflake.Parse(dm.ChannelID)
	if err != nil {
		logger.Error("Invalid channel ID in DutyManager", "channelID", dm.ChannelID, "error", err)
		return
	}
	msgSnowflake, err := snowflake.Parse(*dm.MessageID)
	if err != nil {
		logger.Error("Invalid message ID in DutyManager", "messageID", *dm.MessageID, "error", err)
		return
	}
	embeds := []discord.Embed{embed}
	components := []discord.LayoutComponent{row}
	if _, err := client.Rest.UpdateMessage(chanID, msgSnowflake, discord.MessageUpdate{
		Embeds:     &embeds,
		Components: &components,
	}); err != nil {
		logger.Error("Error editing duty message", "error", err)
	}

	updateBotPresence(client)
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
