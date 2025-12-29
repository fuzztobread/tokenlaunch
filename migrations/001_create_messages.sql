CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(64) PRIMARY KEY,
    external_id VARCHAR(255),
    author VARCHAR(255),
    username VARCHAR(255),
    content TEXT,
    source VARCHAR(50),
    created_at TIMESTAMP,
    ingested_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_username ON messages(username);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);
