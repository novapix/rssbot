CREATE TABLE IF NOT EXISTS discord_channels (
    id SERIAL PRIMARY KEY,
    channel_id TEXT UNIQUE NOT NULL,
    guild_id TEXT,
    name TEXT
);

CREATE TABLE IF NOT EXISTS telegram_chats (
    id SERIAL PRIMARY KEY,
    chat_id TEXT UNIQUE NOT NULL,
    username TEXT,
    type TEXT
);

CREATE TABLE IF NOT EXISTS feeds (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    format TEXT DEFAULT '**{{.Title}}**\n{{.Link}}',
    active BOOLEAN DEFAULT TRUE,
    last_checked TIMESTAMP DEFAULT NOW(),
    discord_id INTEGER REFERENCES discord_channels(id) ON DELETE CASCADE,
    telegram_id INTEGER REFERENCES telegram_chats(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sent_items (
    id SERIAL PRIMARY KEY,
    feed_id INTEGER REFERENCES feeds(id) ON DELETE CASCADE,
    guid TEXT NOT NULL,
    title TEXT,
    sent_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(feed_id, guid)
);
