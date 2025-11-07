package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/novapix/rssbot/logger"
	"github.com/novapix/rssbot/poller"
)

func (b *Bot) StartPoller(p *poller.Poller) {
	go p.Start(func(item poller.RSSItem) {
		channelID := item.Meta["discord_channel"]
		if channelID == "" {
			return
		}

		embed := &discordgo.MessageEmbed{
			Title:       item.Title,
			Description: item.Description,
			Color:       0x00ff00,
			Timestamp:   time.Now().Format(time.RFC3339),
			Fields:      item.Fields,
		}

		if _, err := b.Session.ChannelMessageSendEmbed(channelID, embed); err != nil {
			logger.Error.Printf("Failed to send embed: %v", err)
		}
	})
}
