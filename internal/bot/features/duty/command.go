package duty

import (
	"log"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"

	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "duty",
		Description: "Gérer le gestionnaire de service",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Ajouter un gestionnaire de service",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "duty",
						Description: "Rôle de service",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "oncall",
						Description: "Rôle de semi service",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "offradio",
						Description: "Rôle off radio",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "logchannel",
						Description: "Canal de logs",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Supprimer un gestionnaire de service",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "messageid",
						Description: "ID du message du gestionnaire",
						Required:    true,
					},
				},
			},
		},
	},
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}
	sub := data.Options[0]
	switch sub.Name {
	case "add":
		handleAdd(s, i, sub)
	case "remove":
		handleRemove(s, i, sub)
	}
}

func handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) {
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionManageChannels == 0 {
		respondEphemeral(s, i, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	opts := optionMap(sub.Options)
	dutyRole := opts["duty"].RoleValue(s, i.GuildID)
	onCallRole := opts["oncall"].RoleValue(s, i.GuildID)
	offRadioRole := opts["offradio"].RoleValue(s, i.GuildID)
	logChannel := opts["logchannel"].ChannelValue(s)

	botMember, err := s.GuildMember(i.GuildID, s.State.User.ID)
	if err != nil {
		respondEphemeral(s, i, "Erreur lors de la récupération des informations du bot.")
		return
	}

	guild, err := s.Guild(i.GuildID)
	if err != nil {
		respondEphemeral(s, i, "Erreur lors de la récupération du serveur.")
		return
	}

	botHighest := highestRolePosition(guild.Roles, botMember.Roles)

	for _, r := range []*discordgo.Role{dutyRole, onCallRole, offRadioRole} {
		if r.Position >= botHighest {
			respondEphemeral(s, i, "Un ou plusieurs rôles sont au-dessus ou au même niveau que le rôle le plus haut du bot.")
			return
		}
	}

	members, err := s.GuildMembers(i.GuildID, "", 1000)
	if err != nil {
		respondEphemeral(s, i, "Erreur lors de la récupération des membres.")
		return
	}

	onDuty := membersWithRole(members, dutyRole.ID)
	onCall := membersWithRole(members, onCallRole.ID)
	offRadio := membersWithRole(members, offRadioRole.ID)

	embed, row := BuildDutyEmbed(onDuty, onCall, offRadio)

	msg, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{row},
	})
	if err != nil {
		log.Printf("Error sending duty message: %v", err)
		respondEphemeral(s, i, "Erreur lors de l'envoi du message.")
		return
	}

	dutyRoleID := dutyRole.ID
	onCallRoleID := onCallRole.ID
	offRadioRoleID := offRadioRole.ID
	logsChannelID := logChannel.ID
	msgID := msg.ID

	dm := models.DutyManager{
		GuildID:        i.GuildID,
		ChannelID:      i.ChannelID,
		MessageID:      &msgID,
		DutyRoleID:     &dutyRoleID,
		OnCallRoleID:   &onCallRoleID,
		OffRadioRoleID: &offRadioRoleID,
		LogsChannelID:  &logsChannelID,
	}

	if err := database.DB.Create(&dm).Error; err != nil {
		log.Printf("Error saving DutyManager: %v", err)
		respondEphemeral(s, i, "Erreur lors de l'enregistrement en base de données.")
		return
	}

	respondEphemeral(s, i, "Gestionnaire de service créé avec succès.")
}

func handleRemove(s *discordgo.Session, i *discordgo.InteractionCreate, sub *discordgo.ApplicationCommandInteractionDataOption) {
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionManageChannels == 0 {
		respondEphemeral(s, i, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	messageID := sub.Options[0].StringValue()

	var dm models.DutyManager
	if err := database.DB.Where("guild_id = ? AND message_id = ?", i.GuildID, messageID).First(&dm).Error; err != nil {
		respondEphemeral(s, i, "Gestionnaire de service introuvable.")
		return
	}

	if err := s.ChannelMessageDelete(dm.ChannelID, messageID); err != nil {
		log.Printf("Error deleting duty message: %v", err)
	}

	if err := database.DB.Delete(&dm).Error; err != nil {
		log.Printf("Error deleting DutyManager: %v", err)
		respondEphemeral(s, i, "Erreur lors de la suppression.")
		return
	}

	respondEphemeral(s, i, "Gestionnaire de service supprimé avec succès.")
}

func optionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range opts {
		m[opt.Name] = opt
	}
	return m
}

func membersWithRole(members []*discordgo.Member, roleID string) []string {
	var result []string
	for _, m := range members {
		for _, r := range m.Roles {
			if r == roleID {
				result = append(result, m.User.ID)
				break
			}
		}
	}
	return result
}

func highestRolePosition(roles []*discordgo.Role, memberRoleIDs []string) int {
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

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("Error responding to interaction: %v", err)
	}
}
