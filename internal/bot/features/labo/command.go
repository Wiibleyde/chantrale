package labo

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

var Commands = []*discordgo.ApplicationCommand{
	{
		Name:        "labo",
		Description: "Effectuer une analyse de laboratoire",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "bloodgroup",
				Description: "Analyse de groupe sanguin",
				Options:     commonLaboOptions(bloodgroupChoices()),
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "alcohole",
				Description: "Analyse de taux d'alcoolémie",
				Options:     commonLaboOptions(alcoholeChoices()),
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "drugs",
				Description: "Analyse de dépistage de drogues",
				Options:     drugsOptions(),
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "diseases",
				Description: "Analyse de maladies",
				Options:     commonLaboOptions(diseasesChoices()),
			},
		},
	},
}

func commonLaboOptions(resultChoices []*discordgo.ApplicationCommandOptionChoice) []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "nom_prenom",
			Description: "Nom et prénom du patient",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "resultat",
			Description: "Résultat de l'analyse (optionnel)",
			Required:    false,
			Choices:     resultChoices,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "time",
			Description: "Durée de l'analyse en minutes (optionnel, 1-3 chiffres)",
			Required:    false,
		},
	}
}

func drugsOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "nom_prenom",
			Description: "Nom et prénom du patient",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "depistage",
			Description: "Type de dépistage",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Cannabis", Value: "Cannabis"},
				{Name: "Cocaïne", Value: "Cocaïne"},
				{Name: "Héroïne", Value: "Héroïne"},
				{Name: "Amphétamines", Value: "Amphétamines"},
				{Name: "Ecstasy", Value: "Ecstasy"},
				{Name: "Méthamphétamine", Value: "Méthamphétamine"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "resultat",
			Description: "Résultat de l'analyse (optionnel)",
			Required:    false,
			Choices:     drugsChoices(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "time",
			Description: "Durée de l'analyse en minutes (optionnel, 1-3 chiffres)",
			Required:    false,
		},
	}
}

func bloodgroupChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "O+", Value: "O+"},
		{Name: "A+", Value: "A+"},
		{Name: "B+", Value: "B+"},
		{Name: "AB+", Value: "AB+"},
		{Name: "O-", Value: "O-"},
		{Name: "A-", Value: "A-"},
		{Name: "B-", Value: "B-"},
		{Name: "AB-", Value: "AB-"},
	}
}

func alcoholeChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Négatif", Value: "Négatif"},
		{Name: "Faible", Value: "Faible"},
		{Name: "Moyen", Value: "Moyen"},
		{Name: "Élevé", Value: "Élevé"},
	}
}

func drugsChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Négatif", Value: "Négatif"},
		{Name: "Positif", Value: "Positif"},
	}
}

func diseasesChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "Négatif", Value: "Négatif"},
	}
}

func HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}
	sub := data.Options[0]

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Error("Error deferring", "error", err)
		return
	}

	opts := optionMap(sub.Options)

	patientName := opts["nom_prenom"].StringValue()

	analyseType := sub.Name
	if sub.Name == "drugs" {
		if depistage, ok := opts["depistage"]; ok {
			analyseType = depistage.StringValue()
		}
	}

	analyseTime := defaultTime(sub.Name)
	if timeOpt, ok := opts["time"]; ok {
		if t, err := strconv.Atoi(timeOpt.StringValue()); err == nil {
			if t > 0 && t < 999 {
				analyseTime = t
			}
		}
	}

	var result string
	if resOpt, ok := opts["resultat"]; ok {
		result = resOpt.StringValue()
	} else {
		result = randomResult(sub.Name)
	}

	entry := &LaboEntry{
		ChannelID: i.ChannelID,
		UserID:    i.Member.User.ID,
		StartTime: time.Now(),
		Name:      patientName,
		Type:      analyseType,
		Result:    result,
		Time:      analyseTime,
	}

	waitEmbed := BuildLaboWaitingEmbed(entry)

	msg, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{waitEmbed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Annuler",
						Style:    discordgo.DangerButton,
						CustomID: "laboCancelButton",
					},
				},
			},
		},
	})
	if err != nil {
		logger.Error("Error sending labo message", "error", err)
		if _, err2 := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Erreur lors de l'envoi du message d'analyse.",
			Flags:   discordgo.MessageFlagsEphemeral,
		}); err2 != nil {
			logger.Error("Error creating followup", "error", err2)
		}
		return
	}

	entry.MessageID = msg.ID
	Queue.Add(entry)

	if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Analyse lancée. Résultat dans %d minute(s).", analyseTime),
		Flags:   discordgo.MessageFlagsEphemeral,
	}); err != nil {
		logger.Error("Error creating followup", "error", err)
	}
}

func defaultTime(subCmd string) int {
	switch subCmd {
	case "bloodgroup":
		return 5
	case "alcohole":
		return 3
	case "drugs":
		return 5
	case "diseases":
		return 10
	default:
		return 5
	}
}

func randomResult(subCmd string) string {
	type weighted struct {
		value  string
		weight int
	}

	var pool []weighted

	switch subCmd {
	case "bloodgroup":
		pool = []weighted{
			{"O+", 34}, {"A+", 28}, {"B+", 20}, {"AB+", 2},
			{"O-", 7}, {"A-", 6}, {"B-", 2}, {"AB-", 1},
		}
	case "alcohole":
		pool = []weighted{
			{"Négatif", 90}, {"Faible", 7}, {"Moyen", 2}, {"Élevé", 1},
		}
	case "drugs":
		pool = []weighted{
			{"Négatif", 90}, {"Positif", 10},
		}
	case "diseases":
		return "Négatif"
	default:
		return "Négatif"
	}

	total := 0
	for _, w := range pool {
		total += w.weight
	}

	r := rand.Intn(total)
	cumulative := 0
	for _, w := range pool {
		cumulative += w.weight
		if r < cumulative {
			return w.value
		}
	}

	return pool[len(pool)-1].value
}

func optionMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range opts {
		m[opt.Name] = opt
	}
	return m
}
