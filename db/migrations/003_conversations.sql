-- Migration 003: Conversations
-- Stores serialised agent conversation history keyed by session ID.
-- The history column holds a JSON array of Message objects matching
-- the agent.Message type — the full Anthropic API message history
-- including tool_use and tool_result blocks.
--
-- One row per session. Upserted on every turn so the table self-maintains.
-- Older sessions are pruned by the cleanup query (see server).
-- ============================================================

CREATE TABLE IF NOT EXISTS conversations (
    session_id   TEXT    PRIMARY KEY,
    name         TEXT    NOT NULL,
    history      TEXT    NOT NULL DEFAULT '[]',  -- JSON: []agent.Message
    message_count INTEGER NOT NULL DEFAULT 0,    -- turns (user messages only)
    created_at   TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_conversations_updated
    ON conversations(updated_at);

CREATE TRIGGER IF NOT EXISTS trg_conversations_updated_at
    AFTER UPDATE ON conversations
    BEGIN
        UPDATE conversations SET updated_at = datetime('now') WHERE session_id = NEW.session_id;
    END;
