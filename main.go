package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/novapix/rssbot/db"
	"github.com/novapix/rssbot/discord"
	"github.com/novapix/rssbot/logger"
	"github.com/novapix/rssbot/poller"
)

func main() {
	logger.Init("rssbot.log")

	if err := godotenv.Load(); err != nil {
		logger.Info.Println("No .env file found, using environment variables")
	} else {
		logger.Info.Println(".env loaded successfully")
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		logger.Error.Println("POSTGRES_URL not set in environment")
		os.Exit(1)
	}

	dbConn := db.InitDB(postgresURL)
	logger.Info.Println("Database initialized and schema verified successfully")

	discordToken := os.Getenv("DISCORD_TOKEN")
	discordOwner := os.Getenv("DISCORD_OWNER_ID")
	if discordToken == "" || discordOwner == "" {
		logger.Error.Println("Discord credentials missing in environment")
		os.Exit(1)
	}

	bot, err := discord.NewBot(discordToken, discordOwner, dbConn)
	if err != nil {
		logger.Error.Fatalf("Failed to create Discord bot: %v", err)
	}

	if err := bot.Start(); err != nil {
		logger.Error.Fatalf("Discord bot exited with error: %v", err)
	}

	p := poller.NewPoller(dbConn, 2*time.Minute)
	bot.StartPoller(p)

	logger.Info.Println("Bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	logger.Info.Println("Shutting down...")
	bot.Session.Close()
	dbConn.Close()
	logger.Info.Println("Bot exited successfully")
}
