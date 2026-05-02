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
	APIPort      string
	APIKey       string
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
		APIPort:      os.Getenv("API_PORT"),
		APIKey:       os.Getenv("API_KEY"),
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
