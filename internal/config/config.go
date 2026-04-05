package config

import "os"

type Config struct {
	DiscordToken string
	GuildIDs     []string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
}

func Load() *Config {
	guildIDsStr := os.Getenv("GUILD_IDS")
	var guildIDs []string
	for _, id := range splitCSV(guildIDsStr) {
		if id != "" {
			guildIDs = append(guildIDs, id)
		}
	}
	return &Config{
		DiscordToken: os.Getenv("DISCORD_TOKEN"),
		GuildIDs:     guildIDs,
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       os.Getenv("DB_PORT"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
	}
}

func splitCSV(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	result = append(result, current)
	return result
}
