-- Migration 004: Seed Conversations
-- MySQL version
-- ============================================================
USE nilspcarlson_calendar;
INSERT INTO conversations (session_id, name, history, message_count)
VALUES (
    'nilspcarlson_calendar_0',
    'main_calendar',
    JSON_VALID('[]'),
    0
);
