package mortuary

import (
	"fmt"
	"sort"
	"strings"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"
	"LsmsBot/internal/stats"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "mortuary",
		Description: "Gestion de la morgue",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "init",
				Description: "Initialiser le panneau de la morgue",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Supprimer le panneau de la morgue",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{Name: "messageid", Description: "ID du message du panneau", Required: true},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "add",
				Description: "Ajouter un corps à un casier",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionInt{
						Name:        "casier",
						Description: "Numéro du casier",
						Required:    true,
						Choices:     lockerChoices(),
					},
					discord.ApplicationCommandOptionString{Name: "nom_prenom", Description: "Nom et prénom du défunt", Required: true},
					discord.ApplicationCommandOptionString{Name: "commentaire", Description: "Commentaire optionnel", Required: false},
				},
			},
		},
	},
}

func lockerChoices() []discord.ApplicationCommandOptionChoiceInt {
	choices := make([]discord.ApplicationCommandOptionChoiceInt, 12)
	for i := 0; i < 12; i++ {
		choices[i] = discord.ApplicationCommandOptionChoiceInt{
			Name:  fmt.Sprintf("Casier %d", i+1),
			Value: i + 1,
		}
	}
	return choices
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	data := e.SlashCommandInteractionData()
	if data.SubCommandName == nil {
		return
	}
	switch *data.SubCommandName {
	case "init":
		handleInit(e)
	case "remove":
		handleRemove(e)
	case "add":
		handleAdd(e)
	}
}

func handleInit(e *events.ApplicationCommandInteractionCreate) {
	guildID := *e.GuildID()

	var existing []models.MortuaryManager
	if err := database.DB.Where("guild_id = ?", guildID.String()).Limit(1).Find(&existing).Error; err != nil {
		logger.Error("Error checking mortuary manager", "error", err)
		respondEphemeral(e, "Erreur lors de la vérification du panneau.")
		return
	}
	if len(existing) > 0 {
		respondEphemeral(e, "Un panneau de la morgue existe déjà dans ce serveur.")
		return
	}

	embed := BuildMortuaryEmbed(nil)
	channelID := e.Channel().ID()

	msg, err := e.Client().Rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		logger.Error("Error sending mortuary panel message", "error", err)
		respondEphemeral(e, "Erreur lors de l'envoi du panneau de la morgue.")
		return
	}

	mm := models.MortuaryManager{
		GuildID:   guildID.String(),
		ChannelID: channelID.String(),
		MessageID: msg.ID.String(),
	}
	if err := database.DB.Create(&mm).Error; err != nil {
		logger.Error("Error saving mortuary manager", "error", err)
		respondEphemeral(e, "Erreur lors de la sauvegarde en base de données.")
		return
	}

	stats.Record(guildID.String(), e.Member().User.ID.String(), "mortuary.panel_init", map[string]any{
		"channel_id": channelID.String(),
	})

	respondEphemeral(e, "Panneau de la morgue initialisé avec succès.")
}

func handleAdd(e *events.ApplicationCommandInteractionCreate) {
	guildID := *e.GuildID()

	var mm models.MortuaryManager
	if err := database.DB.Where("guild_id = ?", guildID.String()).First(&mm).Error; err != nil {
		respondEphemeral(e, "Aucun panneau de la morgue trouvé. Utilisez `/mortuary init` d'abord.")
		return
	}

	data := e.SlashCommandInteractionData()
	lockerNumber := data.Int("casier")
	name := data.String("nom_prenom")

	var comment *string
	if v, ok := data.OptString("commentaire"); ok && v != "" {
		comment = &v
	}

	var existingAssignments []models.MortuaryAssignment
	if err := database.DB.Where("guild_id = ? AND locker_number = ?", guildID.String(), lockerNumber).Limit(1).Find(&existingAssignments).Error; err != nil {
		logger.Error("Error checking mortuary assignment", "error", err)
		respondEphemeral(e, "Erreur lors de la vérification du casier.")
		return
	}
	if len(existingAssignments) > 0 {
		respondEphemeral(e, fmt.Sprintf("Le casier %d est déjà occupé par ||%s||.", lockerNumber, existingAssignments[0].Name))
		return
	}

	assignment := models.MortuaryAssignment{
		GuildID:      guildID.String(),
		LockerNumber: int(lockerNumber),
		Name:         name,
		Comment:      comment,
	}
	if err := database.DB.Create(&assignment).Error; err != nil {
		logger.Error("Error creating mortuary assignment", "error", err)
		respondEphemeral(e, "Erreur lors de l'ajout du corps.")
		return
	}

	if err := updateMortuaryPanel(e.Client(), mm); err != nil {
		logger.Error("Error updating mortuary panel", "error", err)
		respondEphemeral(e, "Corps ajouté mais erreur lors de la mise à jour du panneau.")
		return
	}

	stats.Record(guildID.String(), e.Member().User.ID.String(), "mortuary.assign", map[string]any{
		"locker_number": lockerNumber,
		"name":          name,
		"comment":       comment,
	})

	respondEphemeral(e, fmt.Sprintf("Corps de **||%s||** ajouté au casier **%d**.", name, lockerNumber))
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

	var mm models.MortuaryManager
	if err := database.DB.Where("guild_id = ? AND message_id = ?", guildID.String(), messageID).First(&mm).Error; err != nil {
		respondEphemeral(e, "Panneau de la morgue introuvable.")
		return
	}

	chanID, err := snowflake.Parse(mm.ChannelID)
	if err == nil {
		msgSnowflake, err2 := snowflake.Parse(messageID)
		if err2 == nil {
			if err3 := e.Client().Rest.DeleteMessage(chanID, msgSnowflake); err3 != nil {
				logger.Error("Error deleting mortuary panel message", "error", err3)
			}
		}
	}

	if err := database.DB.Where("guild_id = ?", guildID.String()).Delete(&models.MortuaryAssignment{}).Error; err != nil {
		logger.Error("Error deleting mortuary assignments", "error", err)
	}

	if err := database.DB.Delete(&mm).Error; err != nil {
		logger.Error("Error deleting MortuaryManager", "error", err)
		respondEphemeral(e, "Erreur lors de la suppression.")
		return
	}

	stats.Record(guildID.String(), member.User.ID.String(), "mortuary.panel_remove", map[string]any{
		"channel_id": mm.ChannelID,
	})

	respondEphemeral(e, "Panneau de la morgue supprimé avec succès.")
}

func updateMortuaryPanel(client *bot.Client, mm models.MortuaryManager) error {
	var assignments []models.MortuaryAssignment
	if err := database.DB.Where("guild_id = ?", mm.GuildID).Find(&assignments).Error; err != nil {
		return err
	}

	embed := BuildMortuaryEmbed(assignments)
	components := buildMortuaryButtons(assignments)
	embedsList := []discord.Embed{embed}

	chanID, err := snowflake.Parse(mm.ChannelID)
	if err != nil {
		return err
	}
	msgID, err := snowflake.Parse(mm.MessageID)
	if err != nil {
		return err
	}

	_, err = client.Rest.UpdateMessage(chanID, msgID, discord.MessageUpdate{
		Embeds:     &embedsList,
		Components: &components,
	})
	return err
}

func buildMortuaryButtons(assignments []models.MortuaryAssignment) []discord.LayoutComponent {
	if len(assignments) == 0 {
		return []discord.LayoutComponent{}
	}

	sort.Slice(assignments, func(i, j int) bool {
		return assignments[i].LockerNumber < assignments[j].LockerNumber
	})

	var buttons []discord.InteractiveComponent
	for _, a := range assignments {
		label := fmt.Sprintf("Casier %d - %s", a.LockerNumber, a.Name)
		if len(label) > 80 {
			label = label[:80]
		}

		buttons = append(buttons, discord.ButtonComponent{
			Label:    label,
			Style:    discord.ButtonStyleSecondary,
			CustomID: fmt.Sprintf("lsmsMortuary--%d", a.LockerNumber),
		})
	}

	var rows []discord.LayoutComponent
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, discord.ActionRowComponent{Components: buttons[i:end]})
	}
	return rows
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

func lockerFromCustomID(customID string) string {
	parts := strings.SplitN(customID, "--", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
