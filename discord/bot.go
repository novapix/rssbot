package discord

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/novapix/rssbot/logger"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
	OwnerID string
}

func NewBot(token, ownerID string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentMessageContent
	bot := &Bot{
		Session: dg,
		OwnerID: ownerID,
	}

	dg.AddHandler(bot.messageCreate)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return err
	}

	logger.Info.Println("Discord bot is now running")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	b.Session.Close()
	logger.Info.Println("Discord bot stopped")
	return nil
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		logger.Info.Printf("Responded to !ping from %s", m.Author.Username)
	}
}
