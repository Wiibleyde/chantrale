package bed

import (
	"bytes"
	"fmt"
	"strings"

	"LsmsBot/internal/database"
	"LsmsBot/internal/database/models"
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "beds",
		Description: "Gestion des lits de l'hôpital",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "init",
				Description: "Initialiser le panneau des lits",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Ajouter un patient à un lit",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "lit",
						Description: "Lit à attribuer",
						Required:    true,
						Choices:     bedChoices(),
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "nom_prenom",
						Description: "Nom et prénom du patient",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "garde_a_vue",
						Description: "Le patient est-il en garde à vue ?",
						Required:    false,
					},
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "deces",
						Description: "Le patient est-il décédé ?",
						Required:    false,
					},
				},
			},
		},
	},
}

func bedChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(BedConfigs))
	for i, bed := range BedConfigs {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  "Lit " + bed.Letter,
			Value: bed.Letter,
		}
	}
	return choices
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}
	sub := data.Options[0]
	switch sub.Name {
	case "init":
		handleInit(s, i)
	case "add":
		handleAdd(s, i, sub.Options)
	}
}

func handleInit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var existing []models.BedManager
	if err := database.DB.Where("guild_id = ?", i.GuildID).Limit(1).Find(&existing).Error; err != nil {
		logger.Error("Error checking bed manager", "error", err)
		respondEphemeral(s, i, "Erreur lors de la vérification du panneau.")
		return
	}
	if len(existing) > 0 {
		respondEphemeral(s, i, "Un panneau des lits existe déjà dans ce serveur.")
		return
	}

	imgBytes, err := GenerateBedImage(nil)
	if err != nil {
		logger.Error("Error generating bed image", "error", err)
		respondEphemeral(s, i, "Erreur lors de la génération de l'image des lits.")
		return
	}

	embed := BuildBedEmbed()

	msg, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:        "beds.png",
				ContentType: "image/png",
				Reader:      bytes.NewReader(imgBytes),
			},
		},
	})
	if err != nil {
		logger.Error("Error sending bed panel message", "error", err)
		respondEphemeral(s, i, "Erreur lors de l'envoi du panneau des lits.")
		return
	}

	bm := models.BedManager{
		GuildID:   i.GuildID,
		ChannelID: i.ChannelID,
		MessageID: msg.ID,
	}
	if err := database.DB.Create(&bm).Error; err != nil {
		logger.Error("Error saving bed manager", "error", err)
		respondEphemeral(s, i, "Erreur lors de la sauvegarde en base de données.")
		return
	}

	respondEphemeral(s, i, "Panneau des lits initialisé avec succès.")
}

func handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
	var bm models.BedManager
	if err := database.DB.Where("guild_id = ?", i.GuildID).First(&bm).Error; err != nil {
		respondEphemeral(s, i, "Aucun panneau des lits trouvé. Utilisez `/beds init` d'abord.")
		return
	}

	om := optionMap(opts)
	bedLetter := om["lit"].StringValue()
	patientName := om["nom_prenom"].StringValue()

	var underArrest, death bool
	if gavOpt, ok := om["garde_a_vue"]; ok {
		underArrest = gavOpt.BoolValue()
	}
	if decesOpt, ok := om["deces"]; ok {
		death = decesOpt.BoolValue()
	}

	var existingAssignments []models.BedAssignment
	if err := database.DB.Where("guild_id = ? AND bed_letter = ?", i.GuildID, bedLetter).Limit(1).Find(&existingAssignments).Error; err != nil {
		logger.Error("Error checking bed assignment", "error", err)
		respondEphemeral(s, i, "Erreur lors de la vérification du lit.")
		return
	}
	if len(existingAssignments) > 0 {
		respondEphemeral(s, i, fmt.Sprintf("Le lit %s est déjà occupé par %s.", bedLetter, existingAssignments[0].Name))
		return
	}

	assignment := models.BedAssignment{
		GuildID:     i.GuildID,
		BedLetter:   bedLetter,
		Name:        patientName,
		UnderArrest: underArrest,
		Death:       death,
	}
	if err := database.DB.Create(&assignment).Error; err != nil {
		logger.Error("Error creating bed assignment", "error", err)
		respondEphemeral(s, i, "Erreur lors de l'ajout du patient.")
		return
	}

	if err := updateBedPanel(s, bm); err != nil {
		logger.Error("Error updating bed panel", "error", err)
		respondEphemeral(s, i, "Patient ajouté mais erreur lors de la mise à jour du panneau.")
		return
	}

	respondEphemeral(s, i, fmt.Sprintf("Patient **%s** ajouté au lit **%s**.", patientName, bedLetter))
}

func updateBedPanel(s *discordgo.Session, bm models.BedManager) error {
	var assignments []models.BedAssignment
	if err := database.DB.Where("guild_id = ?", bm.GuildID).Find(&assignments).Error; err != nil {
		return err
	}

	imgBytes, err := GenerateBedImage(assignments)
	if err != nil {
		return err
	}

	embed := BuildBedEmbed()
	components := buildBedButtons(assignments)

	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: bm.ChannelID,
		ID:      bm.MessageID,
		Embeds:  &[]*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:        "beds.png",
				ContentType: "image/png",
				Reader:      bytes.NewReader(imgBytes),
			},
		},
		Attachments: &[]*discordgo.MessageAttachment{},
		Components:  &components,
	})
	return err
}

func buildBedButtons(assignments []models.BedAssignment) []discordgo.MessageComponent {
	if len(assignments) == 0 {
		return []discordgo.MessageComponent{}
	}

	var buttons []discordgo.MessageComponent
	for _, a := range assignments {
		label := fmt.Sprintf("Lit %s - %s", a.BedLetter, a.Name)
		if len(label) > 80 {
			label = label[:80]
		}

		style := discordgo.SecondaryButton
		if a.UnderArrest {
			style = discordgo.DangerButton
		}
		if a.Death {
			style = discordgo.DangerButton
		}

		buttons = append(buttons, discordgo.Button{
			Label:    label,
			Style:    style,
			CustomID: "lsmsBed--" + a.BedLetter,
		})
	}

	var rows []discordgo.MessageComponent
	for i := 0; i < len(buttons); i += 5 {
		end := i + 5
		if end > len(buttons) {
			end = len(buttons)
		}
		rows = append(rows, discordgo.ActionsRow{
			Components: buttons[i:end],
		})
	}
	return rows
}

func respondEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Error("Error responding to interaction", "error", err)
	}
}

func optionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range opts {
		m[opt.Name] = opt
	}
	return m
}

func bedLetterFromCustomID(customID string) string {
	parts := strings.SplitN(customID, "--", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
