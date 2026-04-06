package bot

import (
	"os"
	"os/signal"
	"syscall"

	"LsmsBot/internal/bot/embeds"
	"LsmsBot/internal/bot/features/bed"
	"LsmsBot/internal/bot/features/doctor"
	"LsmsBot/internal/bot/features/duty"
	"LsmsBot/internal/bot/features/labo"
	"LsmsBot/internal/bot/features/radio"
	"LsmsBot/internal/bot/router"
	"LsmsBot/internal/config"
	"LsmsBot/internal/logger"

	"github.com/bwmarrin/discordgo"
)

func Run() {
	cfg := config.Load()

	s, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		logger.Fatal("Error creating Discord session", "error", err)
	}

	s.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	labo.Queue.SetSession(s)

	s.AddHandler(func(s *discordgo.Session, e *discordgo.GuildMemberUpdate) {
		duty.HandleGuildMemberUpdate(s, e)
	})

	// --- Register features ---
	// To add a new feature: import its package and call feature.Register(r) here.
	r := router.New()
	duty.Register(r)
	radio.Register(r)
	doctor.Register(r)
	labo.Register(r)
	bed.Register(r)
	r.Attach(s)

	// Collect all slash commands from each feature
	allCommands := []*discordgo.ApplicationCommand{}
	allCommands = append(allCommands, duty.Commands...)
	allCommands = append(allCommands, radio.Commands...)
	allCommands = append(allCommands, doctor.Commands...)
	allCommands = append(allCommands, labo.Commands...)
	allCommands = append(allCommands, bed.Commands...)

	s.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		logger.Info("Logged in", "user", s.State.User.Username+"#"+s.State.User.Discriminator)
		for _, guildID := range cfg.GuildIDs {
			for _, cmd := range allCommands {
				if _, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd); err != nil {
					logger.Error("Cannot create command", "command", cmd.Name, "guild", guildID, "error", err)
				}
			}
			logger.Info("Commands registered", "guild", guildID)
		}
	})

	if err := s.Open(); err != nil {
		logger.Fatal("Error opening connection", "error", err)
	}
	defer s.Close()

	embeds.Init(s)
	duty.StartScheduler(s)

	logger.Info("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Sending duty summary before shutdown...")
	duty.SendShutdownSummary(s)

	logger.Info("Shutting down...")
}
