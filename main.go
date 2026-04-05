package main

import (
	"LsmsBot/internal/bot"
	"LsmsBot/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.Init()
	bot.Run()
}
