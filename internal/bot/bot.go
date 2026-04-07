package bot

import (
	"context"
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

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

func Run() {
	cfg := config.Load()

	client, err := disgo.New(cfg.DiscordToken,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMembers),
		),
	)
	if err != nil {
		logger.Fatal("Error creating Discord client", "error", err)
	}

	labo.Queue.SetClient(client)

	client.AddEventListeners(bot.NewListenerFunc(duty.HandleGuildMemberUpdate))

	r := router.New()
	duty.Register(r)
	radio.Register(r)
	doctor.Register(r)
	labo.Register(r)
	bed.Register(r)
	r.Attach(client)

	allCommands := []discord.ApplicationCommandCreate{}
	allCommands = append(allCommands, duty.Commands...)
	allCommands = append(allCommands, radio.Commands...)
	allCommands = append(allCommands, doctor.Commands...)
	allCommands = append(allCommands, labo.Commands...)
	allCommands = append(allCommands, bed.Commands...)

	client.AddEventListeners(bot.NewListenerFunc(func(e *events.Ready) {
		logger.Info("Logged in", "user", e.User.Username)
		embeds.Init(e.User.EffectiveAvatarURL())
		duty.StartScheduler(client)
		for _, guildID := range cfg.GuildIDs {
			guildSnowflake := snowflake.MustParse(guildID)
			if _, err := e.Client().Rest.SetGuildCommands(e.Application.ID, guildSnowflake, allCommands); err != nil {
				logger.Error("Cannot set commands", "guild", guildID, "error", err)
			}
		}
		logger.Info("Commands registered")
	}))

	ctx := context.Background()
	if err := client.OpenGateway(ctx); err != nil {
		logger.Fatal("Error opening connection", "error", err)
	}
	defer client.Close(ctx)

	logger.Info("Bot is running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Sending duty summary before shutdown...")
	duty.SendShutdownSummary(client)

	logger.Info("Shutting down...")
}
