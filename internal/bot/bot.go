package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"LsmsBot/internal/bot/features/doctor"
	"LsmsBot/internal/bot/features/duty"
	"LsmsBot/internal/bot/features/labo"
	"LsmsBot/internal/bot/features/radio"
	"LsmsBot/internal/bot/router"
	"LsmsBot/internal/config"

	"github.com/bwmarrin/discordgo"
)

func Run() {
	cfg := config.Load()

	s, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
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
	r.Attach(s)

	// Collect all slash commands from each feature
	allCommands := []*discordgo.ApplicationCommand{}
	allCommands = append(allCommands, duty.Commands...)
	allCommands = append(allCommands, radio.Commands...)
	allCommands = append(allCommands, doctor.Commands...)
	allCommands = append(allCommands, labo.Commands...)

	s.AddHandler(func(s *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
		for _, guildID := range cfg.GuildIDs {
			for _, cmd := range allCommands {
				if _, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd); err != nil {
					log.Printf("Cannot create '%v' command in guild %v: %v", cmd.Name, guildID, err)
				}
			}
			log.Printf("Commands registered in guild %s", guildID)
		}
	})

	if err := s.Open(); err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}
	defer s.Close()

	log.Println("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Shutting down...")
}
