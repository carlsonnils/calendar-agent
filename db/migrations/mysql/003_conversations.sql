-- Migration 003: Conversations
-- MySQL version
-- ============================================================

CREATE TABLE IF NOT EXISTS conversations (
    session_id    VARCHAR(255)  NOT NULL,
    name          VARCHAR(255)  NOT NULL,
    history       JSON          NOT NULL,
    message_count INT           NOT NULL DEFAULT 0,
    created_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (session_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE INDEX idx_conversations_updated ON conversations(updated_at);
