package discord

import (
	"database/sql"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/novapix/rssbot/discord/commands"
	"github.com/novapix/rssbot/logger"
)

type Bot struct {
	Session      *discordgo.Session
	OwnerID      string
	DB           *sql.DB
	addFeedState map[string]*commands.AddFeedSession
	commands     map[string]commands.CommandHandler
}

func NewBot(token, ownerID string, dbConn *sql.DB) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	bot := &Bot{
		Session:      dg,
		OwnerID:      ownerID,
		DB:           dbConn,
		addFeedState: make(map[string]*commands.AddFeedSession),
		commands:     make(map[string]commands.CommandHandler),
	}

	bot.registerCommands()
	dg.AddHandler(bot.messageCreate)

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return err
	}
	logger.Info.Println("✅ Discord bot is now running")
	return nil
}

// Build context for commands
func (b *Bot) buildContext() commands.BotContext {
	return commands.BotContext{
		DB:      b.DB,
		OwnerID: b.OwnerID,
		SendMsg: func(channelID, content string) error {
			_, err := b.Session.ChannelMessageSend(channelID, content)
			return err
		},
	}
}

func (b *Bot) registerCommands() {
	ctx := b.buildContext()

	b.commands["!ping"] = commands.CommandHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		commands.PingCommand(s, m, ctx)
	})

	b.commands["!rssadd"] = commands.CommandHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		commands.RSSAddStart(s, m, ctx, b.addFeedState)
	})
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.TrimSpace(m.Content)
	userID := m.Author.ID

	// Execute command if matches
	for cmd, handler := range b.commands {
		if strings.HasPrefix(content, cmd) {
			handler(s, m)
			return
		}
	}

	// Check if user is in an active RSSAdd session
	if session, ok := b.addFeedState[userID]; ok {
		commands.RSSAddStep(s, m, b.buildContext(), session)
	}
}
