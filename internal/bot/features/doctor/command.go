package doctor

import (
	"fmt"

	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

type Formation struct {
	Title       string
	Competences []string
}

var formations = []Formation{
	{
		Title:       "Formations secondaires",
		Competences: []string{"PPA", "Hélicoptère", "Bateau", "Psychiatrie", "Pôle funéraire"},
	},
	{
		Title:       "Stagiaire - Explications",
		Competences: []string{"Visite de l'hôpital", "Visite du bureau psy", "Visite de la morgue", "Brancard", "Mettre / sortir d'un véhicule", "Intranet"},
	},
	{
		Title:       "Stagiaire - Formations",
		Competences: []string{"Conduite de l'ambulance", "Utilisation de la radio", "Réanimation", "Bobologie", "Radiologie / IRM", "Anesthésie", "Opération", "Suture", "Don du sang", "Etat alcoolisé / drogué", "Prise d'otage", "Gestion d'un patient en état d'arrestation", "Autonomie", "Daronora"},
	},
	{
		Title:       "Interne",
		Competences: []string{"Indépendance", "Fiche patient.es", "Rapports médicaux", "Communication sur la radio LSMS/LSPD", "Supervision de stagiaires", "Opérations avancées", "Visite médicale / Certificat médical"},
	},
}

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "doctor",
		Description: "Créer un dossier de formation médical",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:         discordgo.ApplicationCommandOptionChannel,
				Name:         "forumchannel",
				Description:  "Canal forum pour le dossier",
				Required:     true,
				ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildForum},
			},
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Utilisateur concerné",
				Required:    true,
			},
		},
	},
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	perms, err := s.UserChannelPermissions(i.Member.User.ID, i.ChannelID)
	if err != nil || perms&discordgo.PermissionManageChannels == 0 {
		respondEphemeral(s, i, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Error("Error deferring", "error", err)
		return
	}

	data := i.ApplicationCommandData()
	opts := optionMap(data.Options)

	forumChannel := opts["forumchannel"].ChannelValue(s)
	user := opts["user"].UserValue(s)

	displayName := user.Username
	if member, err := s.GuildMember(i.GuildID, user.ID); err == nil && member.Nick != "" {
		displayName = member.Nick
	}

	threadName := fmt.Sprintf("Dossier de formation - %s", displayName)
	initialEmbed := BuildDossierEmbed(displayName)

	// ForumThreadStartComplex returns (*Channel, error)
	thread, err := s.ForumThreadStartComplex(forumChannel.ID, &discordgo.ThreadStart{
		Name: threadName,
	}, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{initialEmbed},
	})
	if err != nil {
		logger.Error("Error creating forum thread", "error", err)
		if _, err2 := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Erreur lors de la création du fil de discussion.",
			Flags:   discordgo.MessageFlagsEphemeral,
		}); err2 != nil {
			logger.Error("Error creating followup", "error", err2)
		}
		return
	}

	for _, f := range formations {
		embed := BuildFormationEmbed(f.Title)
		if _, err := s.ChannelMessageSendComplex(thread.ID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{embed},
		}); err != nil {
			logger.Error("Error sending formation embed", "error", err)
			continue
		}
		for _, competence := range f.Competences {
			if _, err := s.ChannelMessageSend(thread.ID, fmt.Sprintf("- %s", competence)); err != nil {
				logger.Error("Error sending competence message", "error", err)
			}
		}
	}

	if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Dossier de formation créé avec succès dans <#%s>.", thread.ID),
		Flags:   discordgo.MessageFlagsEphemeral,
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}

func optionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range opts {
		m[opt.Name] = opt
	}
	return m
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
