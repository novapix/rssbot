package poller

import (
	"database/sql"
	"strings"
	"time"

	"github.com/novapix/rssbot/logger"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

type RSSItem struct {
	Title       string
	Description string
	Fields      []*discordgo.MessageEmbedField
	Meta        map[string]string
}

type Poller struct {
	DB       *sql.DB
	Interval time.Duration
}

func NewPoller(db *sql.DB, interval time.Duration) *Poller {
	return &Poller{
		DB:       db,
		Interval: interval,
	}
}

func (p *Poller) Start(callback func(item RSSItem)) {
	ticker := time.NewTicker(p.Interval)
	defer ticker.Stop()
	logger.Info.Printf("Poller started, interval: %v", p.Interval)

	for range ticker.C {
		if err := p.checkFeeds(callback); err != nil {
			logger.Error.Printf("Poller error: %v", err)
		}
	}
}

func (p *Poller) checkFeeds(callback func(item RSSItem)) error {
	rows, err := p.DB.Query(`SELECT id, url, format, discord_id FROM feeds WHERE active = TRUE`)
	if err != nil {
		return err
	}
	defer rows.Close()

	fp := gofeed.NewParser()

	for rows.Next() {
		var feedID int
		var url, format string
		var discordID sql.NullInt64

		if err := rows.Scan(&feedID, &url, &format, &discordID); err != nil {
			logger.Error.Printf("Failed to scan feed: %v", err)
			continue
		}

		meta := make(map[string]string)
		if discordID.Valid {
			var channelID string
			err := p.DB.QueryRow(`SELECT channel_id FROM discord_channels WHERE id=$1`, discordID.Int64).Scan(&channelID)
			if err != nil {
				logger.Error.Printf("Failed to get channel_id for feed %d: %v", feedID, err)
				continue
			}
			meta["discord_channel"] = channelID
		}

		p.processFeed(feedID, url, format, meta, callback, fp)
	}

	return nil
}

func (p *Poller) processFeed(feedID int, url, format string, meta map[string]string, callback func(item RSSItem), fp *gofeed.Parser) {
	feed, err := fp.ParseURL(url)
	if err != nil {
		logger.Error.Printf("Failed to parse feed %s: %v", url, err)
		return
	}

	for _, item := range feed.Items {
		var exists bool
		err := p.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM sent_items WHERE feed_id=$1 AND guid=$2)`, feedID, item.GUID).Scan(&exists)
		if err != nil {
			logger.Error.Printf("Failed to check sent_items: %v", err)
			continue
		}

		if exists {
			continue
		}

		title, description, fields := replacePlaceholdersWithFields(format, item)
		callback(RSSItem{Title: title, Description: description, Fields: fields, Meta: meta})

		_, err = p.DB.Exec(`INSERT INTO sent_items (feed_id, guid, title) VALUES ($1, $2, $3)`, feedID, item.GUID, item.Title)
		if err != nil {
			logger.Error.Printf("Failed to insert sent_item: %v", err)
		}
	}
}

// func replacePlaceholdersWithFields(format string, item *gofeed.Item) (title, description string, fields []*discordgo.MessageEmbedField) {
// 	title = item.Title
// 	description = ""
// 	fields = []*discordgo.MessageEmbedField{}

// 	lines := strings.Split(format, "\n")
// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.HasPrefix(line, "Field:") {
// 			parts := strings.SplitN(line[6:], "=", 2)
// 			if len(parts) == 2 {
// 				name := strings.TrimSpace(parts[0])
// 				value := replacePlaceholdersInString(parts[1], item)
// 				fields = append(fields, &discordgo.MessageEmbedField{Name: name, Value: value, Inline: true})
// 			}
// 		} else {
// 			description += replacePlaceholdersInString(line, item) + "\n"
// 		}
// 	}

//		description = strings.TrimSpace(description)
//		return
//	}
func replacePlaceholdersWithFields(format string, item *gofeed.Item) (title, description string, fields []*discordgo.MessageEmbedField) {
	// Title will be empty; user format controls description and fields
	title = ""

	description = ""
	fields = []*discordgo.MessageEmbedField{}

	lines := strings.Split(format, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Field:") {
			parts := strings.SplitN(line[6:], "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value := replacePlaceholdersInString(parts[1], item)
				fields = append(fields, &discordgo.MessageEmbedField{Name: name, Value: value, Inline: true})
			}
		} else {
			description += replacePlaceholdersInString(line, item) + "\n"
		}
	}

	description = strings.TrimSpace(description)
	return
}

func replacePlaceholdersInString(str string, item *gofeed.Item) string {
	placeholders := map[string]string{
		"Title":       item.Title,
		"Link":        item.Link,
		"Description": item.Description,
		"Published":   "",
		"Author":      "",
	}

	if item.PublishedParsed != nil {
		placeholders["Published"] = item.PublishedParsed.Format("2006-01-02 15:04:05")
	}
	if item.Author != nil {
		placeholders["Author"] = item.Author.Name
	}

	for ns, ext := range item.Extensions {
		for key, vals := range ext {
			if len(vals) > 0 {
				placeholderName := ns
				if key != "" {
					placeholderName += "." + key
				}
				placeholders[placeholderName] = vals[0].Value
			}
		}
	}

	for k, v := range placeholders {
		str = strings.ReplaceAll(str, "{{."+k+"}}", v)
	}

	return str
}
