package doctor

import (
	"fmt"

	"LsmsBot/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
)

type Formation struct {
	Title       string
	Competences []string
}

var formations = []Formation{
	{
		Title:       "Formations secondaires",
		Competences: []string{"PPA", "Hélicoptère", "Bateau", "Psychiatrie", "Pôle funéraire", "Dispatcher"},
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

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name:        "doctor",
		Description: "Créer un dossier de formation médical",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:         "forumchannel",
				Description:  "Canal forum pour le dossier",
				Required:     true,
				ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildForum},
			},
			discord.ApplicationCommandOptionUser{
				Name:        "user",
				Description: "Utilisateur concerné",
				Required:    true,
			},
		},
	},
}

func HandleCommand(e *events.ApplicationCommandInteractionCreate) {
	member := e.Member()
	if member == nil || !member.Permissions.Has(discord.PermissionManageChannels) {
		respondEphemeral(e, "Vous n'avez pas la permission de gérer les canaux.")
		return
	}

	if err := e.DeferCreateMessage(true); err != nil {
		logger.Error("Error deferring", "error", err)
		return
	}

	data := e.SlashCommandInteractionData()
	forumChannel, _ := data.OptChannel("forumchannel")
	user, _ := data.OptUser("user")

	client := e.Client()
	guildID := *e.GuildID()

	displayName := user.Username
	if guildMember, err := client.Rest.GetMember(guildID, user.ID); err == nil && guildMember.Nick != nil {
		displayName = *guildMember.Nick
	}

	threadName := fmt.Sprintf("Dossier de formation - %s", displayName)
	initialEmbed := BuildDossierEmbed(displayName)

	thread, err := client.Rest.CreatePostInThreadChannel(forumChannel.ID, discord.ThreadChannelPostCreate{
		Name: threadName,
		Message: discord.MessageCreate{
			Embeds: []discord.Embed{initialEmbed},
		},
	})
	if err != nil {
		logger.Error("Error creating forum thread", "error", err)
		if _, err2 := client.Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
			Content: "Erreur lors de la création du fil de discussion.",
			Flags:   discord.MessageFlagEphemeral,
		}); err2 != nil {
			logger.Error("Error creating followup", "error", err2)
		}
		return
	}

	for _, f := range formations {
		embed := BuildFormationEmbed(f.Title)
		if _, err := client.Rest.CreateMessage(thread.ID(), discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		}); err != nil {
			logger.Error("Error sending formation embed", "error", err)
			continue
		}
		for _, competence := range f.Competences {
			if _, err := client.Rest.CreateMessage(thread.ID(), discord.MessageCreate{
				Content: fmt.Sprintf("- %s", competence),
			}); err != nil {
				logger.Error("Error sending competence message", "error", err)
			}
		}
	}

	if _, err := client.Rest.CreateFollowupMessage(e.ApplicationID(), e.Token(), discord.MessageCreate{
		Content: fmt.Sprintf("Dossier de formation créé avec succès dans <#%s>.", thread.ID()),
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
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
