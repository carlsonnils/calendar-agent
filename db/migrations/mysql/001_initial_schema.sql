-- Migration 001: Initial Schema
-- Calendar Agent - Personal Ops Database
-- MySQL version
-- ============================================================

-- ============================================================
-- PROJECTS
-- ============================================================
CREATE TABLE IF NOT EXISTS projects (
    id          INT           NOT NULL AUTO_INCREMENT,
    name        VARCHAR(255)  NOT NULL,
    description TEXT,
    status      ENUM('active', 'paused', 'completed', 'archived')
                              NOT NULL DEFAULT 'active',
    color       VARCHAR(7),                         -- hex e.g. '#4A90E2'
    due_date    DATE,                               -- e.g. '2026-12-31'
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- EVENTS
-- ============================================================
CREATE TABLE IF NOT EXISTS events (
    id          INT           NOT NULL AUTO_INCREMENT,
    title       VARCHAR(255)  NOT NULL,
    description TEXT,
    start_at    DATETIME      NOT NULL,
    end_at      DATETIME      NOT NULL,
    location    VARCHAR(255),
    all_day     TINYINT(1)    NOT NULL DEFAULT 0,   -- 0=false, 1=true
    project_id  INT,
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT chk_events_all_day   CHECK (all_day IN (0, 1)),
    CONSTRAINT chk_events_end_after CHECK (end_at >= start_at),
    CONSTRAINT fk_events_project    FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- TASKS
-- ============================================================
CREATE TABLE IF NOT EXISTS tasks (
    id          INT           NOT NULL AUTO_INCREMENT,
    title       VARCHAR(255)  NOT NULL,
    description TEXT,
    status      ENUM('todo', 'in_progress', 'blocked', 'done', 'cancelled')
                              NOT NULL DEFAULT 'todo',
    priority    TINYINT       NOT NULL DEFAULT 2,   -- 1=high, 2=medium, 3=low
    due_at      DATETIME,
    project_id  INT,
    parent_id   INT,
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT chk_tasks_priority CHECK (priority IN (1, 2, 3)),
    CONSTRAINT fk_tasks_project   FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT fk_tasks_parent    FOREIGN KEY (parent_id)
        REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- REMINDERS
-- ============================================================
CREATE TABLE IF NOT EXISTS reminders (
    id               INT           NOT NULL AUTO_INCREMENT,
    message          TEXT          NOT NULL,
    remind_at        DATETIME      NOT NULL,

    event_id         INT,
    task_id          INT,

    recurrence_type  ENUM('minutely', 'hourly', 'daily', 'weekly', 'monthly', 'yearly', 'custom'),
    -- JSON for complex rules, e.g.:
    --   {"interval": 2}                              every 2 weeks (with recurrence_type='weekly')
    --   {"interval": 1, "days": ["mon","wed","fri"]} specific weekdays
    --   {"interval": 1, "day_of_month": 15}          15th of every month
    recurrence_rule  JSON,

    recurrence_end   DATE,

    is_active        TINYINT(1)    NOT NULL DEFAULT 1,
    last_fired_at    DATETIME,
    created_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT chk_reminders_is_active     CHECK (is_active IN (0, 1)),
    CONSTRAINT chk_reminders_single_attach CHECK (event_id IS NULL OR task_id IS NULL),
    CONSTRAINT fk_reminders_event          FOREIGN KEY (event_id)
        REFERENCES events(id) ON DELETE CASCADE,
    CONSTRAINT fk_reminders_task           FOREIGN KEY (task_id)
        REFERENCES tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- LOG / DIARY / THOUGHT RECORDS
-- ============================================================
CREATE TABLE IF NOT EXISTS log_entries (
    id          INT           NOT NULL AUTO_INCREMENT,

    entry_type  ENUM('note', 'diary', 'work_log', 'thought', 'mood', 'idea', 'win', 'gratitude')
                              NOT NULL DEFAULT 'note',

    title       VARCHAR(255),
    body        TEXT          NOT NULL,

    mood_score  TINYINT       CHECK (mood_score BETWEEN 1 AND 100),

    -- CBT thought record fields
    situation         TEXT,
    automatic_thought TEXT,
    reframe           TEXT,

    project_id  INT,
    task_id     INT,

    -- tags stored as JSON array: ["health","work","personal"]
    -- queryable via JSON_TABLE() or JSON_CONTAINS()
    tags        JSON          NOT NULL,

    entry_date  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    CONSTRAINT fk_log_project FOREIGN KEY (project_id)
        REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT fk_log_task    FOREIGN KEY (task_id)
        REFERENCES tasks(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX idx_events_start   ON events(start_at);
CREATE INDEX idx_events_project ON events(project_id);

CREATE INDEX idx_tasks_status  ON tasks(status);
CREATE INDEX idx_tasks_due     ON tasks(due_at);
CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_parent  ON tasks(parent_id);

-- Daemon polls this every minute — keep it tight
CREATE INDEX idx_reminders_next_fire ON reminders(remind_at, is_active);

CREATE INDEX idx_log_entry_date  ON log_entries(entry_date);
CREATE INDEX idx_log_entry_type  ON log_entries(entry_type);
CREATE INDEX idx_log_mood_score  ON log_entries(mood_score);
CREATE INDEX idx_log_project     ON log_entries(project_id);
CREATE INDEX idx_log_task        ON log_entries(task_id);

-- ============================================================
-- NOTE: updated_at triggers are replaced by
-- ON UPDATE CURRENT_TIMESTAMP column defaults above.
-- No explicit triggers are required in MySQL for this purpose.
-- ============================================================
