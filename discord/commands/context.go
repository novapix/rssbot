package commands

import "github.com/bwmarrin/discordgo"

type BotContext struct {
	DB      interface{} // pass *sql.DB
	OwnerID string
	SendMsg func(channelID, content string) error
}

type CommandHandler func(s *discordgo.Session, m *discordgo.MessageCreate)
