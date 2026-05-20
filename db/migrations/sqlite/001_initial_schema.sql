-- Migration 001: Initial Schema
-- Calendar Agent - Personal Ops Database
-- ============================================================

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- ============================================================
-- PROJECTS
-- Containers for grouping related tasks and events.
-- ============================================================
CREATE TABLE IF NOT EXISTS projects (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    description TEXT,
    status      TEXT    NOT NULL DEFAULT 'active'
                        CHECK(status IN ('active', 'paused', 'completed', 'archived')),
    color       TEXT,                               -- hex e.g. '#4A90E2', for WhatsApp formatting
    due_date    TEXT,                               -- ISO8601 date: '2026-12-31'
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- EVENTS
-- Discrete calendar entries with a start and end time.
-- Optionally linked to a project.
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT    NOT NULL,
    description TEXT,
    start_at    TEXT    NOT NULL,                   -- ISO8601 datetime: '2026-05-01T14:00:00'
    end_at      TEXT    NOT NULL,                   -- ISO8601 datetime: '2026-05-01T15:00:00'
    location    TEXT,
    all_day     INTEGER NOT NULL DEFAULT 0
                        CHECK(all_day IN (0, 1)),   -- boolean: 0=false, 1=true
    project_id  INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now')),

    CHECK(end_at >= start_at)
);

-- ============================================================
-- TASKS
-- Todo items with optional project membership and subtask
-- support via parent_id self-reference.
-- ============================================================
CREATE TABLE IF NOT EXISTS tasks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    title       TEXT    NOT NULL,
    description TEXT,
    status      TEXT    NOT NULL DEFAULT 'todo'
                        CHECK(status IN ('todo', 'in_progress', 'blocked', 'done', 'cancelled')),
    priority    INTEGER NOT NULL DEFAULT 2
                        CHECK(priority IN (1, 2, 3)),   -- 1=high, 2=medium, 3=low
    due_at      TEXT,                               -- ISO8601 datetime, optional
    project_id  INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    parent_id   INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- REMINDERS
-- Scheduled, recurring, or indefinite notifications.
-- Can be standalone or attached to an event or task.
-- remind_at always holds the NEXT fire time; the daemon
-- advances it after each fire for recurring reminders.
-- ============================================================
CREATE TABLE IF NOT EXISTS reminders (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    message          TEXT    NOT NULL,
    remind_at        TEXT    NOT NULL,              -- ISO8601 datetime: next fire time

    -- optional attachment (both NULL = standalone reminder)
    event_id         INTEGER REFERENCES events(id) ON DELETE CASCADE,
    task_id          INTEGER REFERENCES tasks(id)  ON DELETE CASCADE,

    -- recurrence (all NULL = one-shot)
    recurrence_type  TEXT    CHECK(recurrence_type IN (
                         'minutely', 'hourly', 'daily', 'weekly', 'monthly', 'yearly', 'custom'
                     )),
    -- JSON for complex rules, e.g.:
    --   {"interval": 2}                              every 2 weeks (with recurrence_type='weekly')
    --   {"interval": 1, "days": ["mon","wed","fri"]} specific weekdays
    --   {"interval": 1, "day_of_month": 15}          15th of every month
    recurrence_rule  TEXT,

    recurrence_end   TEXT,                          -- ISO8601 date, NULL = fires indefinitely

    -- state
    is_active        INTEGER NOT NULL DEFAULT 1
                             CHECK(is_active IN (0, 1)),
    last_fired_at    TEXT,
    created_at       TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT    NOT NULL DEFAULT (datetime('now')),

    -- a reminder can be attached to at most one thing
    CHECK(
        (event_id IS NULL OR task_id IS NULL)
    )
);

-- ============================================================
-- LOG / DIARY / THOUGHT RECORDS
-- Freeform journal entries. Supports multiple entry types so
-- the same table serves as a diary, a thought record (CBT-style),
-- a work log, a quick capture inbox, and a mood tracker.
--
-- The agent can write entries on your behalf
-- ("log that I finished the report") or you can dictate
-- freeform ("diary entry: today was rough because...")
-- ============================================================
CREATE TABLE IF NOT EXISTS log_entries (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,

    -- type drives how the agent interprets and surfaces entries
    entry_type  TEXT    NOT NULL DEFAULT 'note'
                        CHECK(entry_type IN (
                            'note',         -- general freeform thought or observation
                            'diary',        -- narrative journal / end-of-day entry
                            'work_log',     -- what I worked on, accomplishments, blockers
                            'thought',      -- CBT-style thought record (uses mood/reframe cols)
                            'mood',         -- quick mood snapshot
                            'idea',         -- capture for later, not yet a task
                            'win',          -- positive reinforcement / accomplishment log
                            'gratitude'     -- gratitude entry
                        )),

    title       TEXT,                               -- optional headline; agent can auto-generate
    body        TEXT    NOT NULL,                   -- the actual content

    -- mood / emotional context (optional)
    -- integer 1-100 so the agent can query trends: "how has my mood been this month?"
    mood_score  INTEGER CHECK(mood_score BETWEEN 1 AND 100),

    -- CBT thought record fields (optional, most useful for entry_type = 'thought')
    -- kept as columns rather than JSON so the agent can query them individually
    -- e.g. "show me all my reframes from last month"
    situation         TEXT,                         -- what happened / the trigger
    automatic_thought TEXT,                         -- the immediate negative thought
    reframe           TEXT,                         -- the rational / balanced reframe

    -- optional linkage to other entities
    project_id  INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    task_id     INTEGER REFERENCES tasks(id)    ON DELETE SET NULL,

    -- tags as JSON array: ["health","work","personal"]
    -- queryable via SQLite json_each():
    --   SELECT * FROM log_entries, json_each(tags) WHERE json_each.value = 'health'
    tags        TEXT    NOT NULL DEFAULT '[]',

    -- when the entry is logically "about" — defaults to now but can be backdated
    -- e.g. "log yesterday's standup" sets entry_date to yesterday
    entry_date  TEXT    NOT NULL DEFAULT (datetime('now')),

    created_at  TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- ============================================================
-- INDEXES
-- ============================================================

-- events: range queries by start time are the dominant pattern
CREATE INDEX IF NOT EXISTS idx_events_start
    ON events(start_at);

CREATE INDEX IF NOT EXISTS idx_events_project
    ON events(project_id)
    WHERE project_id IS NOT NULL;

-- tasks
CREATE INDEX IF NOT EXISTS idx_tasks_status
    ON tasks(status);

CREATE INDEX IF NOT EXISTS idx_tasks_due
    ON tasks(due_at)
    WHERE due_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_project
    ON tasks(project_id)
    WHERE project_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tasks_parent
    ON tasks(parent_id)
    WHERE parent_id IS NOT NULL;

-- reminders: daemon polls this every minute — keep it as tight as possible
CREATE INDEX IF NOT EXISTS idx_reminders_next_fire
    ON reminders(remind_at)
    WHERE is_active = 1;

-- log entries: date range and type are the most common query axes
CREATE INDEX IF NOT EXISTS idx_log_entry_date
    ON log_entries(entry_date);

CREATE INDEX IF NOT EXISTS idx_log_entry_type
    ON log_entries(entry_type);

CREATE INDEX IF NOT EXISTS idx_log_mood_score
    ON log_entries(mood_score)
    WHERE mood_score IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_log_project
    ON log_entries(project_id)
    WHERE project_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_log_task
    ON log_entries(task_id)
    WHERE task_id IS NOT NULL;

-- ============================================================
-- UPDATED_AT TRIGGERS
-- SQLite has no ON UPDATE equivalent — triggers keep
-- updated_at accurate without relying on the application layer.
-- ============================================================

CREATE TRIGGER IF NOT EXISTS trg_projects_updated_at
    AFTER UPDATE ON projects
    BEGIN
        UPDATE projects SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS trg_events_updated_at
    AFTER UPDATE ON events
    BEGIN
        UPDATE events SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS trg_tasks_updated_at
    AFTER UPDATE ON tasks
    BEGIN
        UPDATE tasks SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS trg_reminders_updated_at
    AFTER UPDATE ON reminders
    BEGIN
        UPDATE reminders SET updated_at = datetime('now') WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS trg_log_entries_updated_at
    AFTER UPDATE ON log_entries
    BEGIN
        UPDATE log_entries SET updated_at = datetime('now') WHERE id = NEW.id;
    END;
