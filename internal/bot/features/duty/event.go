package duty

import (
	"log"
	"sync"
	"time"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"

	"github.com/bwmarrin/discordgo"
)

var (
	debounceMap = make(map[string]*time.Timer)
	debounceMu  sync.Mutex
)

func HandleGuildMemberUpdate(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
	key := e.GuildID + "-" + e.Member.User.ID

	debounceMu.Lock()
	if t, ok := debounceMap[key]; ok {
		t.Stop()
	}

	var prevRoles []string
	if oldMember, err := s.State.Member(e.GuildID, e.Member.User.ID); err == nil {
		prevRoles = oldMember.Roles
	}

	debounceMap[key] = time.AfterFunc(500*time.Millisecond, func() {
		debounceMu.Lock()
		delete(debounceMap, key)
		debounceMu.Unlock()

		handleMemberRoleChange(s, e.GuildID, e.Member, prevRoles)
	})
	debounceMu.Unlock()
}

func handleMemberRoleChange(s *discordgo.Session, guildID string, member *discordgo.Member, prevRoles []string) {
	newRoles := member.Roles
	added := diffRoles(prevRoles, newRoles)
	removed := diffRoles(newRoles, prevRoles)
	changed := append(added, removed...)

	if len(changed) == 0 {
		return
	}

	var dms []models.DutyManager
	if err := database.DB.Where("guild_id = ?", guildID).Find(&dms).Error; err != nil {
		log.Printf("Error fetching DutyManagers: %v", err)
		return
	}

	members, err := s.GuildMembers(guildID, "", 1000)
	if err != nil {
		log.Printf("Error fetching members: %v", err)
		return
	}

	for _, dm := range dms {
		affected := false
		for _, roleID := range changed {
			if (dm.DutyRoleID != nil && *dm.DutyRoleID == roleID) ||
				(dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleID) ||
				(dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleID) {
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
			onDuty = membersWithRole(members, *dm.DutyRoleID)
		}
		if dm.OnCallRoleID != nil {
			onCall = membersWithRole(members, *dm.OnCallRoleID)
		}
		if dm.OffRadioRoleID != nil {
			offRadio = membersWithRole(members, *dm.OffRadioRoleID)
		}

		embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)
		components := []discordgo.MessageComponent{row}
		if _, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         *dm.MessageID,
			Channel:    dm.ChannelID,
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &components,
		}); err != nil {
			log.Printf("Error editing duty message: %v", err)
		}

		if dm.LogsChannelID == nil {
			continue
		}

		userID := member.User.ID
		for _, roleID := range added {
			var logEmbed *discordgo.MessageEmbed
			if dm.DutyRoleID != nil && *dm.DutyRoleID == roleID {
				logEmbed = BuildDutyUpdateEmbed(userID, true)
			} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleID {
				logEmbed = BuildOnCallUpdateEmbed(userID, true)
			} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleID {
				logEmbed = BuildOffRadioUpdateEmbed(userID, true)
			}
			if logEmbed != nil {
				if _, err := s.ChannelMessageSendComplex(*dm.LogsChannelID, &discordgo.MessageSend{
					Embeds: []*discordgo.MessageEmbed{logEmbed},
				}); err != nil {
					log.Printf("Error sending log embed: %v", err)
				}
			}
		}
		for _, roleID := range removed {
			var logEmbed *discordgo.MessageEmbed
			if dm.DutyRoleID != nil && *dm.DutyRoleID == roleID {
				logEmbed = BuildDutyUpdateEmbed(userID, false)
			} else if dm.OnCallRoleID != nil && *dm.OnCallRoleID == roleID {
				logEmbed = BuildOnCallUpdateEmbed(userID, false)
			} else if dm.OffRadioRoleID != nil && *dm.OffRadioRoleID == roleID {
				logEmbed = BuildOffRadioUpdateEmbed(userID, false)
			}
			if logEmbed != nil {
				if _, err := s.ChannelMessageSendComplex(*dm.LogsChannelID, &discordgo.MessageSend{
					Embeds: []*discordgo.MessageEmbed{logEmbed},
				}); err != nil {
					log.Printf("Error sending log embed: %v", err)
				}
			}
		}
	}
}

func diffRoles(a, b []string) []string {
	set := make(map[string]bool)
	for _, r := range a {
		set[r] = true
	}
	var diff []string
	for _, r := range b {
		if !set[r] {
			diff = append(diff, r)
		}
	}
	return diff
}
