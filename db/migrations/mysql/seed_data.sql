-- Migration 002: Seed Data
-- MySQL version
-- All datetimes use MySQL DATE_ADD(NOW(), INTERVAL ...) so data is
-- always "live" relative to whenever the migration runs.
-- ============================================================

-- ============================================================
-- PROJECTS
-- ============================================================

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (1, 'Home Renovation',
        'Kitchen and bathroom remodel. Contractor starts next month.',
        'active', '#E07B54', DATE_ADD(CURDATE(), INTERVAL 4 MONTH));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (2, 'Work Q2 Deliverables',
        'Client data analysis and reporting for Q2. Snowflake pipeline + Power BI dashboards.',
        'active', '#4A90E2', DATE_ADD(CURDATE(), INTERVAL 2 MONTH));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (3, 'Carlson Rankings',
        'Personal web project — leaderboard app. Go backend, SQLite, vanilla JS frontend.',
        'active', '#7ED321', NULL);

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (4, 'Golf Season Prep',
        'Equipment decisions, practice plan, and tournament schedule for the season.',
        'active', '#9B59B6', DATE_ADD(CURDATE(), INTERVAL 28 DAY));

INSERT INTO projects (id, name, description, status, color, due_date)
VALUES (5, 'Pi Home Server',
        'Raspberry Pi 5 VPS hardening, services, and self-hosted tooling.',
        'paused', '#F39C12', NULL);

-- ============================================================
-- EVENTS
-- ============================================================

-- 3 days from now at 10:00
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (1, 'Contractor Walkthrough',
        'Initial walkthrough with renovation contractor. Bring measurements and mood board.',
        DATE_ADD(DATE(NOW()), INTERVAL (3 * 24 * 60 + 10 * 60) MINUTE),
        DATE_ADD(DATE(NOW()), INTERVAL (3 * 24 * 60 + 11 * 60 + 30) MINUTE),
        '123 Main St', 0, 1);

-- 4 days from now at 14:00
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (2, 'Q2 Client Check-in',
        'Bi-weekly sync with client. Review pipeline progress and dashboard drafts.',
        DATE_ADD(DATE(NOW()), INTERVAL (4 * 24 * 60 + 14 * 60) MINUTE),
        DATE_ADD(DATE(NOW()), INTERVAL (4 * 24 * 60 + 15 * 60) MINUTE),
        'Google Meet', 0, 2);

-- 6 days from now at 07:30
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (3, 'Golf Round - Pease Golf Course',
        'Morning round. Tee time booked. Testing the new shaft setup.',
        DATE_ADD(DATE(NOW()), INTERVAL (6 * 24 * 60 + 7 * 60 + 30) MINUTE),
        DATE_ADD(DATE(NOW()), INTERVAL (6 * 24 * 60 + 12 * 60) MINUTE),
        'Pease Golf Course, Portsmouth NH', 0, 4);

-- 9 days from now at 09:00
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (4, 'Dentist Appointment',
        'Routine checkup and cleaning.',
        DATE_ADD(DATE(NOW()), INTERVAL (9 * 24 * 60 + 9 * 60) MINUTE),
        DATE_ADD(DATE(NOW()), INTERVAL (9 * 24 * 60 + 10 * 60) MINUTE),
        'Downtown Dental, 45 Market St', 0, NULL);

-- 21 days from now — all day event
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (5, 'Team Off-site',
        NULL,
        DATE_ADD(DATE(NOW()), INTERVAL 21 DAY),
        DATE_ADD(DATE_ADD(DATE(NOW()), INTERVAL 21 DAY), INTERVAL 1439 MINUTE),
        NULL, 1, NULL);

-- Tonight at 20:00
INSERT INTO events (id, title, description, start_at, end_at, location, all_day, project_id)
VALUES (6, 'Pi Server Maintenance Window',
        'Update packages, review CrowdSec logs, rotate WireGuard keys.',
        DATE_ADD(DATE(NOW()), INTERVAL 20 HOUR),
        DATE_ADD(DATE(NOW()), INTERVAL (20 * 60 + 90) MINUTE),
        NULL, 0, 5);

-- ============================================================
-- TASKS
-- ============================================================

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (1, 'Get three contractor quotes',
        'Compare quotes for kitchen demo, cabinet install, and plumbing rough-in.',
        'done', 1,
        DATE_ADD(DATE(NOW()), INTERVAL (-3 * 24 * 60 + 17 * 60) MINUTE),
        1, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (2, 'Order kitchen cabinet hardware',
        'Brushed nickel pulls from Liberty Hardware. 32 pieces needed.',
        'todo', 2,
        DATE_ADD(DATE(NOW()), INTERVAL (12 * 24 * 60 + 17 * 60) MINUTE),
        1, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (3, 'Confirm tile delivery date',
        'Porcelain floor tile — 180 sq ft. Supplier is Tile Depot.',
        'in_progress', 1,
        DATE_ADD(DATE(NOW()), INTERVAL (5 * 24 * 60 + 17 * 60) MINUTE),
        1, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (4, 'Build Snowflake ingestion pipeline',
        'Pull source data into Snowflake staging schema. Python + Snowflake connector.',
        'in_progress', 1,
        DATE_ADD(DATE(NOW()), INTERVAL (13 * 24 * 60 + 17 * 60) MINUTE),
        2, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (5, 'Write DAX measures for revenue dashboard',
        NULL,
        'todo', 2,
        DATE_ADD(DATE(NOW()), INTERVAL (19 * 24 * 60 + 17 * 60) MINUTE),
        2, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (6, 'Validate row counts post-pipeline',
        'Cross-check staging vs source. Log discrepancies.',
        'todo', 1,
        DATE_ADD(DATE(NOW()), INTERVAL (14 * 24 * 60 + 17 * 60) MINUTE),
        2, 4);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (7, 'Add CSV export to leaderboard',
        'Export current standings to CSV. Single button in the UI.',
        'todo', 3, NULL, 3, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (8, 'Fix iOS Safari scroll bug in score cards',
        'Score card grid overflows viewport on Safari 17. Likely min-height issue.',
        'in_progress', 1, NULL, 3, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (9, 'Implement dense rank tiebreaker logic',
        'DENSE_RANK window function for tied scores. Already working in SQL — wire to UI.',
        'done', 1, NULL, 3, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (10, 'Decide on iron shaft - X100 vs KBS Tour X',
        'Pull launch monitor numbers. Compare EI curve and kick point. Book range session.',
        'in_progress', 1,
        DATE_ADD(DATE(NOW()), INTERVAL (7 * 24 * 60 + 17 * 60) MINUTE),
        4, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (11, 'Book tee times for next 4 weekends',
        NULL,
        'todo', 2,
        DATE_ADD(DATE(NOW()), INTERVAL (7 * 24 * 60 + 17 * 60) MINUTE),
        4, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (12, 'Set up Postfix SMS relay',
        'Route alerts through Postfix to vtext.com gateway for Verizon.',
        'blocked', 2, NULL, 5, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (13, 'Configure nftables rate limiting',
        'Add rate limit rules to complement CrowdSec bouncer.',
        'todo', 2, NULL, 5, NULL);

INSERT INTO tasks (id, title, description, status, priority, due_at, project_id, parent_id)
VALUES (14, 'Renew car registration',
        NULL,
        'todo', 2,
        DATE_ADD(DATE(NOW()), INTERVAL (28 * 24 * 60 + 17 * 60) MINUTE),
        NULL, NULL);

-- ============================================================
-- REMINDERS
-- ============================================================

INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (1, 'Contractor walkthrough in 1 hour — grab the measurements folder.',
        DATE_ADD(DATE(NOW()), INTERVAL (3 * 24 * 60 + 9 * 60) MINUTE),
        1, NULL, NULL, NULL, NULL, 1);

INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (2, 'Follow up with Tile Depot on delivery date.',
        DATE_ADD(DATE(NOW()), INTERVAL (24 * 60 + 9 * 60) MINUTE),
        NULL, 3, NULL, NULL, NULL, 1);

INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (3, 'Take morning supplements.',
        DATE_ADD(DATE(NOW()), INTERVAL (24 * 60 + 8 * 60) MINUTE),
        NULL, NULL, 'daily', '{"interval": 1}', NULL, 1);

-- Next Friday at 17:00
-- NOTE: SQLite 'weekday 5' (next Friday) is approximated here with a fixed +7 days.
-- For production use, set remind_at to the actual next Friday date at deployment time.
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (4, 'Write your weekly work log — what did you ship, what is blocked?',
        DATE_ADD(DATE_ADD(DATE(NOW()), INTERVAL (7 - WEEKDAY(NOW()) + 4) % 7 DAY), INTERVAL 17 HOUR),
        NULL, NULL, 'weekly', '{"interval": 1, "days": ["fri"]}', NULL, 1);

-- Next Monday at 07:00, ends in ~6 months
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (5, 'Golf practice session this week — book the range if you have not.',
        DATE_ADD(DATE_ADD(DATE(NOW()), INTERVAL (7 - WEEKDAY(NOW())) % 7 DAY), INTERVAL 7 HOUR),
        NULL, NULL, 'weekly', '{"interval": 1, "days": ["mon"]}',
        DATE_ADD(CURDATE(), INTERVAL 180 DAY), 1);

INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (6, 'Dentist appointment tomorrow at 9am — 45 Market St.',
        DATE_ADD(DATE(NOW()), INTERVAL (8 * 24 * 60 + 20 * 60) MINUTE),
        4, NULL, NULL, NULL, NULL, 1);

-- 1st of next month at 09:00
INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (7, 'Monthly Pi server audit — check CrowdSec logs, update packages, review firewall rules.',
        DATE_ADD(DATE_ADD(LAST_DAY(NOW()), INTERVAL 1 DAY), INTERVAL 9 HOUR),
        NULL, NULL, 'monthly', '{"interval": 1, "day_of_month": 1}', NULL, 1);

INSERT INTO reminders (id, message, remind_at, event_id, task_id, recurrence_type, recurrence_rule, recurrence_end, is_active)
VALUES (8, 'Submit Q1 timesheets.',
        DATE_ADD(DATE(NOW()), INTERVAL (-32 * 24 * 60 + 9 * 60) MINUTE),
        NULL, NULL, NULL, NULL, NULL, 0);

-- ============================================================
-- LOG ENTRIES
-- ============================================================

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (1, 'diary', 'End of day',
        'Productive day overall. Got the Snowflake pipeline running against the staging schema — first full load came through clean. Still need to validate row counts but the hard part is done. Went for a walk after dinner, felt good.',
        7, NULL, NULL, NULL, 2, NULL,
        '["work","health"]',
        DATE_ADD(DATE(NOW()), INTERVAL (-24 * 60 + 22 * 60) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (2, 'work_log', 'Snowflake pipeline first load',
        'Completed initial Python ingestion script. Used Snowflake connector with key-pair auth. Staging schema loaded 4 tables. Next: row count validation against source and then ADF upsert pipeline wiring.',
        NULL, NULL, NULL, NULL, 2, 4,
        '["snowflake","python","etl"]',
        DATE_ADD(DATE(NOW()), INTERVAL (-24 * 60 + 17 * 60 + 30) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (3, 'thought', 'Anxiety about client deliverable timeline',
        'Feeling behind on Q2 work. Spent time sitting with it and working through the thought.',
        4,
        'Client check-in is in 4 days and the dashboard DAX measures are not written yet.',
        'I am going to miss the deadline and the client will lose confidence in me.',
        'The pipeline is done which was the hardest part. DAX measures are 1-2 days of work. I have flagged the timeline proactively before — the client appreciates honesty. I can send a quick update today.',
        2, NULL,
        '["anxiety","work","cbt"]',
        DATE_ADD(DATE(NOW()), INTERVAL (8 * 60 + 30) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (4, 'idea', 'Calendar agent — voice memo integration',
        'Could wire a WhatsApp voice message through Whisper transcription before passing to the agent. Would make logging diary entries and quick task captures much faster while driving.',
        NULL, NULL, NULL, NULL, NULL, NULL,
        '["pi","calendar-agent","idea"]',
        DATE_ADD(DATE(NOW()), INTERVAL (9 * 60 + 15) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (5, 'mood', NULL,
        'Feeling solid. Coffee is good. Ready to work.',
        8, NULL, NULL, NULL, NULL, NULL,
        '[]',
        DATE_ADD(DATE(NOW()), INTERVAL (7 * 60 + 45) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (6, 'win', 'Carlson Rankings leaderboard query working',
        'DENSE_RANK window function finally handling tied scores correctly. Took two attempts to get the partition right but it is clean now. Small win but satisfying.',
        NULL, NULL, NULL, NULL, 3, 9,
        '["carlson-rankings","sql","win"]',
        DATE_ADD(DATE(NOW()), INTERVAL (-4 * 24 * 60 + 16 * 60) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (7, 'note', 'Iron shaft decision — notes from range session',
        'Hit X100 and KBS Tour X back to back. X100 felt more consistent on mishits. KBS Tour X had slightly higher launch which could work given my swing speed. Will pull the launch monitor numbers before deciding. Leaning X100.',
        NULL, NULL, NULL, NULL, 4, 10,
        '["golf","equipment"]',
        DATE_ADD(DATE(NOW()), INTERVAL (-5 * 24 * 60 + 13 * 60) MINUTE));

INSERT INTO log_entries (id, entry_type, title, body, mood_score, situation, automatic_thought, reframe, project_id, task_id, tags, entry_date)
VALUES (8, 'gratitude', NULL,
        'Good weather this week. Range session felt productive. Pi server has been rock solid since the CrowdSec setup.',
        7, NULL, NULL, NULL, NULL, NULL,
        '["gratitude"]',
        DATE_ADD(DATE(NOW()), INTERVAL (-24 * 60 + 21 * 60) MINUTE));
