package main

import (
	"flag"

	"LsmsBot/internal/bot"
	"LsmsBot/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	undeployGlobal := flag.Bool("undeploy-global", false, "Supprime toutes les commandes globales Discord et quitte")
	flag.Parse()

	godotenv.Load()

	if *undeployGlobal {
		bot.UndeployGlobalCommands()
		return
	}

	database.Init()
	bot.Run()
}
