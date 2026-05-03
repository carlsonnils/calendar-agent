-- Migration 002: Seed Data
-- All datetimes use SQLite datetime('now', ...) so data is always "live"
-- relative to whenever the migration runs.
--
-- Individual INSERT statements are used throughout (not multi-row VALUES)
-- for reliable parsing by the SQLite CLI across all versions.
--
-- Offset reference:
--   date('now', '+N days')                   -- date only
--   datetime('now', '+N days', 'start of day', '+H hours')  -- datetime at specific hour
--   datetime('now', 'weekday W', ...)         -- 0=Sun,1=Mon...5=Fri,6=Sat
-- ============================================================

-- ============================================================
-- PROJECTS
-- ============================================================

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (1, 'Home Renovation',
        'Kitchen and bathroom remodel. Contractor starts next month.',
        'active', '#E07B54', date('now', '+4 months'));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (2, 'Work Q2 Deliverables',
        'Client data analysis and reporting for Q2. Snowflake pipeline + Power BI dashboards.',
        'active', '#4A90E2', date('now', '+2 months'));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (3, 'Carlson Rankings',
        'Personal web project — leaderboard app. Go backend, SQLite, vanilla JS frontend.',
        'active', '#7ED321', NULL);

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (4, 'Golf Season Prep',
        'Equipment decisions, practice plan, and tournament schedule for the season.',
        'active', '#9B59B6', date('now', '+28 days'));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (5, 'Pi Home Server',
        'Raspberry Pi 5 VPS hardening, services, and self-hosted tooling.',
        'paused', '#F39C12', NULL);

-- ============================================================
-- EVENTS
-- ============================================================

-- 3 days from now at 10:00 — upcoming contractor walkthrough
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (1, 'Contractor Walkthrough',
        'Initial walkthrough with renovation contractor. Bring measurements and mood board.',
        datetime('now', '+3 days', 'start of day', '+10 hours'),
        datetime('now', '+3 days', 'start of day', '+11 hours', '+30 minutes'),
        '123 Main St', 0, 1);

-- 4 days from now at 14:00 — client sync
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (2, 'Q2 Client Check-in',
        'Bi-weekly sync with client. Review pipeline progress and dashboard drafts.',
        datetime('now', '+4 days', 'start of day', '+14 hours'),
        datetime('now', '+4 days', 'start of day', '+15 hours'),
        'Google Meet', 0, 2);

-- 6 days from now at 07:30 — weekend golf round
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (3, 'Golf Round - Pease Golf Course',
        'Morning round. Tee time booked. Testing the new shaft setup.',
        datetime('now', '+6 days', 'start of day', '+7 hours', '+30 minutes'),
        datetime('now', '+6 days', 'start of day', '+12 hours'),
        'Pease Golf Course, Portsmouth NH', 0, 4);

-- 9 days from now at 09:00 — dentist
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (4, 'Dentist Appointment',
        'Routine checkup and cleaning.',
        datetime('now', '+9 days', 'start of day', '+9 hours'),
        datetime('now', '+9 days', 'start of day', '+10 hours'),
        'Downtown Dental, 45 Market St', 0, NULL);

-- 21 days from now — all day event
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (5, 'Team Off-site',
        NULL,
        datetime('now', '+21 days', 'start of day'),
        datetime('now', '+21 days', 'start of day', '+23 hours', '+59 minutes'),
        NULL, 1, NULL);

-- Tonight at 20:00 — same day as migration runs
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (6, 'Pi Server Maintenance Window',
        'Update packages, review CrowdSec logs, rotate WireGuard keys.',
        datetime('now', 'start of day', '+20 hours'),
        datetime('now', 'start of day', '+21 hours', '+30 minutes'),
        NULL, 0, 5);

-- ============================================================
-- TASKS
-- ============================================================

-- Home Renovation — completed 3 days ago
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (1, 'Get three contractor quotes',
        'Compare quotes for kitchen demo, cabinet install, and plumbing rough-in.',
        'done', 1,
        datetime('now', '-3 days', 'start of day', '+17 hours'),
        1, NULL);

-- Home Renovation — due in ~2 weeks
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (2, 'Order kitchen cabinet hardware',
        'Brushed nickel pulls from Liberty Hardware. 32 pieces needed.',
        'todo', 2,
        datetime('now', '+12 days', 'start of day', '+17 hours'),
        1, NULL);

-- Home Renovation — due in 5 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (3, 'Confirm tile delivery date',
        'Porcelain floor tile — 180 sq ft. Supplier is Tile Depot.',
        'in_progress', 1,
        datetime('now', '+5 days', 'start of day', '+17 hours'),
        1, NULL);

-- Work Q2 — due in 13 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (4, 'Build Snowflake ingestion pipeline',
        'Pull source data into Snowflake staging schema. Python + Snowflake connector.',
        'in_progress', 1,
        datetime('now', '+13 days', 'start of day', '+17 hours'),
        2, NULL);

-- Work Q2 — due in 19 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (5, 'Write DAX measures for revenue dashboard',
        NULL,
        'todo', 2,
        datetime('now', '+19 days', 'start of day', '+17 hours'),
        2, NULL);

-- Work Q2 — subtask of task 4, due in 14 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (6, 'Validate row counts post-pipeline',
        'Cross-check staging vs source. Log discrepancies.',
        'todo', 1,
        datetime('now', '+14 days', 'start of day', '+17 hours'),
        2, 4);

-- Carlson Rankings — no due date
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (7, 'Fix iOS Safari score card layout',
        'backdrop-filter not rendering correctly on iOS 17. Needs -webkit- prefix audit.',
        'todo', 2, NULL, 3, NULL);

-- Carlson Rankings — no due date
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (8, 'Add login modal validation',
        'Client-side validation before POST. Show inline errors on empty fields.',
        'in_progress', 2, NULL, 3, NULL);

-- Carlson Rankings — already done
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (9, 'Write DENSE_RANK leaderboard query',
        'Use window function to handle tied scores correctly.',
        'done', 1, NULL, 3, NULL);

-- Golf — due in 3 days (before walkthrough)
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (10, 'Decide on iron shaft - X100 vs KBS Tour X',
        'Review launch monitor data from fitting. Pull strokes gained notes.',
        'in_progress', 1,
        datetime('now', '+3 days', 'start of day', '+17 hours'),
        4, NULL);

-- Golf — due in 7 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (11, 'Book second range session',
        'Confirm with instructor for mid-month slot.',
        'todo', 3,
        datetime('now', '+7 days', 'start of day', '+17 hours'),
        4, NULL);

-- Pi Server — blocked, no due date
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (12, 'Set up Postfix SMS relay',
        'Route alerts through Postfix to vtext.com gateway for Verizon.',
        'blocked', 2, NULL, 5, NULL);

-- Pi Server — no due date
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (13, 'Configure nftables rate limiting',
        'Add rate limit rules to complement CrowdSec bouncer.',
        'todo', 2, NULL, 5, NULL);

-- Standalone — no project, due in 28 days
INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (14, 'Renew car registration',
        NULL,
        'todo', 2,
        datetime('now', '+28 days', 'start of day', '+17 hours'),
        NULL, NULL);

-- ============================================================
-- REMINDERS
-- ============================================================

-- One-shot: 1 hour before contractor walkthrough (event 1, 3 days out at 09:00)
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (1, 'Contractor walkthrough in 1 hour — grab the measurements folder.',
        datetime('now', '+3 days', 'start of day', '+9 hours'),
        1, NULL, NULL, NULL, NULL, 1);

-- One-shot: tomorrow morning nudge on tile delivery task
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (2, 'Follow up with Tile Depot on delivery date.',
        datetime('now', '+1 day', 'start of day', '+9 hours'),
        NULL, 3, NULL, NULL, NULL, 1);

-- Daily recurring: supplements, fires tomorrow morning, no end date
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (3, 'Take morning supplements.',
        datetime('now', '+1 day', 'start of day', '+8 hours'),
        NULL, NULL, 'daily', '{"interval": 1}', NULL, 1);

-- Weekly recurring: end-of-week work log, next Friday at 17:00
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (4, 'Write your weekly work log — what did you ship, what is blocked?',
        datetime('now', 'weekday 5', 'start of day', '+17 hours'),
        NULL, NULL, 'weekly', '{"interval": 1, "days": ["fri"]}', NULL, 1);

-- Weekly recurring: golf practice nudge, next Monday, ends in 6 months
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (5, 'Golf practice session this week — book the range if you have not.',
        datetime('now', 'weekday 1', 'start of day', '+7 hours'),
        NULL, NULL, 'weekly', '{"interval": 1, "days": ["mon"]}',
        date('now', '+180 days'), 1);

-- One-shot: evening before dentist appointment (event 4, 9 days out)
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (6, 'Dentist appointment tomorrow at 9am — 45 Market St.',
        datetime('now', '+8 days', 'start of day', '+20 hours'),
        4, NULL, NULL, NULL, NULL, 1);

-- Monthly recurring: Pi server audit, fires on 1st of next month at 09:00
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (7, 'Monthly Pi server audit — check CrowdSec logs, update packages, review firewall rules.',
        datetime('now', 'start of month', '+1 month', '+9 hours'),
        NULL, NULL, 'monthly', '{"interval": 1, "day_of_month": 1}', NULL, 1);

-- Inactive: already fired ~32 days ago
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (8, 'Submit Q1 timesheets.',
        datetime('now', '-32 days', 'start of day', '+9 hours'),
        NULL, NULL, NULL, NULL, NULL, 0);

-- ============================================================
-- LOG ENTRIES
-- ============================================================

-- Diary: yesterday evening
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (1, 'diary', 'End of day',
        'Productive day overall. Got the Snowflake pipeline running against the staging schema — first full load came through clean. Still need to validate row counts but the hard part is done. Went for a walk after dinner, felt good.',
        7, NULL, NULL, NULL, 2, NULL,
        '["work","health"]',
        datetime('now', '-1 day', 'start of day', '+22 hours'));

-- Work log: yesterday afternoon
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (2, 'work_log', 'Snowflake pipeline first load',
        'Completed initial Python ingestion script. Used Snowflake connector with key-pair auth. Staging schema loaded 4 tables. Next: row count validation against source and then ADF upsert pipeline wiring.',
        NULL, NULL, NULL, NULL, 2, 4,
        '["snowflake","python","etl"]',
        datetime('now', '-1 day', 'start of day', '+17 hours', '+30 minutes'));

-- Thought record: this morning
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (3, 'thought', 'Anxiety about client deliverable timeline',
        'Feeling behind on Q2 work. Spent time sitting with it and working through the thought.',
        4,
        'Client check-in is in 4 days and the dashboard DAX measures are not written yet.',
        'I am going to miss the deadline and the client will lose confidence in me.',
        'The pipeline is done which was the hardest part. DAX measures are 1-2 days of work. I have flagged the timeline proactively before — the client appreciates honesty. I can send a quick update today.',
        2, NULL,
        '["anxiety","work","cbt"]',
        datetime('now', 'start of day', '+8 hours', '+30 minutes'));

-- Idea: this morning
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (4, 'idea', 'Calendar agent — voice memo integration',
        'Could wire a WhatsApp voice message through Whisper transcription before passing to the agent. Would make logging diary entries and quick task captures much faster while driving.',
        NULL, NULL, NULL, NULL, NULL, NULL,
        '["pi","calendar-agent","idea"]',
        datetime('now', 'start of day', '+9 hours', '+15 minutes'));

-- Mood: this morning
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (5, 'mood', NULL,
        'Feeling solid. Coffee is good. Ready to work.',
        8, NULL, NULL, NULL, NULL, NULL,
        '[]',
        datetime('now', 'start of day', '+7 hours', '+45 minutes'));

-- Win: 4 days ago
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (6, 'win', 'Carlson Rankings leaderboard query working',
        'DENSE_RANK window function finally handling tied scores correctly. Took two attempts to get the partition right but it is clean now. Small win but satisfying.',
        NULL, NULL, NULL, NULL, 3, 9,
        '["carlson-rankings","sql","win"]',
        datetime('now', '-4 days', 'start of day', '+16 hours'));

-- Note: 5 days ago
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (7, 'note', 'Iron shaft decision — notes from range session',
        'Hit X100 and KBS Tour X back to back. X100 felt more consistent on mishits. KBS Tour X had slightly higher launch which could work given my swing speed. Will pull the launch monitor numbers before deciding. Leaning X100.',
        NULL, NULL, NULL, NULL, 4, 10,
        '["golf","equipment"]',
        datetime('now', '-5 days', 'start of day', '+13 hours'));

-- Gratitude: yesterday evening
INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (8, 'gratitude', NULL,
        'Good weather this week. Range session felt productive. Pi server has been rock solid since the CrowdSec setup.',
        7, NULL, NULL, NULL, NULL, NULL,
        '["gratitude"]',
        datetime('now', '-1 day', 'start of day', '+21 hours'));
