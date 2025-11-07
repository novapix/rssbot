package commands

import (
	"database/sql"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/novapix/rssbot/logger"
)

type AddFeedSession struct {
	Step      int
	ChannelID string
	URL       string
	Format    string
}

const (
	insertDiscordChannelQuery = `
INSERT INTO discord_channels (channel_id, guild_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (channel_id) DO UPDATE SET name = EXCLUDED.name
RETURNING id
`

	insertFeedQuery = `
INSERT INTO feeds (url, format, discord_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
`
)

func RSSAddStart(s *discordgo.Session, m *discordgo.MessageCreate, ctx BotContext, addFeedState map[string]*AddFeedSession) {
	if m.Author.ID != ctx.OwnerID {
		ctx.SendMsg(m.ChannelID, "❌ You are not authorized to add feeds.")
		return
	}

	addFeedState[m.Author.ID] = &AddFeedSession{Step: 0}
	ctx.SendMsg(m.ChannelID, "Please send the RSS feed URL:")
}

func RSSAddStep(s *discordgo.Session, m *discordgo.MessageCreate, ctx BotContext, session *AddFeedSession) {
	switch session.Step {
	case 0:
		session.URL = m.Content
		session.ChannelID = m.ChannelID
		session.Format = "**{{.Title}}**\n{{.Link}}"
		session.Step = 1
		ctx.SendMsg(m.ChannelID, "Feed URL received. Send a custom format or type 'skip' to use default:")

	case 1:
		if strings.ToLower(m.Content) != "skip" {
			session.Format = m.Content
		}
		registerChannelAndSaveFeed(ctx, session)
	}
}

func registerChannelAndSaveFeed(ctx BotContext, session *AddFeedSession) {
	db, ok := ctx.DB.(*sql.DB)
	if !ok {
		logger.Error.Println("DB is not *sql.DB")
		return
	}

	var discordID int
	err := db.QueryRow(insertDiscordChannelQuery, session.ChannelID, "", "").Scan(&discordID)
	if err != nil {
		ctx.SendMsg(session.ChannelID, "❌ Failed to register channel: "+err.Error())
		logger.Error.Printf("Failed to insert discord channel: %v", err)
		return
	}

	_, err = db.Exec(insertFeedQuery, session.URL, session.Format, discordID)
	if err != nil {
		ctx.SendMsg(session.ChannelID, "❌ Failed to save feed: "+err.Error())
		logger.Error.Printf("Failed to insert feed: %v", err)
	} else {
		ctx.SendMsg(session.ChannelID, "✅ RSS feed added successfully!")
		logger.Info.Printf("Feed added: %s", session.URL)
	}
}
