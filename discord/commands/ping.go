package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/novapix/rssbot/logger"
)

func PingCommand(s *discordgo.Session, m *discordgo.MessageCreate, ctx BotContext) {
	ctx.SendMsg(m.ChannelID, "Pong!")
	logger.Info.Printf("Responded to !ping from %s (%s)", m.Author.Username, m.Author.ID)
}
