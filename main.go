package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/novapix/rssbot/db"
	"github.com/novapix/rssbot/discord"
	"github.com/novapix/rssbot/logger"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Init("rssbot.log")
		logger.Info.Println("No .env file found, using environment variables")
	} else {
		logger.Init("rssbot.log")
		logger.Info.Println(".env loaded successfully")
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		logger.Error.Println("POSTGRES_URL not set in environment")
		os.Exit(1)
	}

	database := db.InitDB(postgresURL)
	if database != nil {
		logger.Info.Println("Database initialized and schema verified successfully")
	}

	discordToken := os.Getenv("DISCORD_TOKEN")
	discordOwner := os.Getenv("DISCORD_OWNER_ID")
	if discordToken == "" || discordOwner == "" {
		logger.Error.Println("Discord credentials missing in .env")
		os.Exit(1)
	}

	bot, err := discord.NewBot(discordToken, discordOwner)
	if err != nil {
		logger.Error.Fatalf("Failed to create Discord bot: %v", err)
	}

	if err := bot.Start(); err != nil {
		logger.Error.Fatalf("Discord bot exited with error: %v", err)
	}
}
