-- Migration 004: Seed Conversations
-- MySQL version
-- ============================================================

INSERT INTO conversations (session_id, name, history, message_count, created_at, updated_at)
VALUES (
    'session-seed-a',
    'whats today?',
    '[
        {
            "role": "user",
            "content": [{"type": "text", "text": "What have I got on today?"}]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_a1",
                    "name": "get_agenda",
                    "input": {"period": "today"}
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_a1",
                    "content": "{\"items\":[{\"kind\":\"event\",\"id\":6,\"title\":\"Pi Server Maintenance Window\",\"at\":\"2026-05-06T20:00:00Z\",\"extra\":\"\"},{\"kind\":\"reminder\",\"id\":3,\"title\":\"Take morning supplements.\",\"at\":\"2026-05-06T08:00:00Z\",\"extra\":\"daily\"}]}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [{"type": "text", "text": "Today you have your morning supplements reminder and the Pi server maintenance window at 8pm. That is it for today — reasonably light."}]
        },
        {
            "role": "user",
            "content": [{"type": "text", "text": "Good. Add a task to review the nftables rate limiting config before the maintenance window, high priority."}]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_a2",
                    "name": "create_task",
                    "input": {"title": "Review nftables rate limiting config", "priority": 1, "project_id": 5, "due_at": "2026-05-06T19:00:00"}
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_a2",
                    "content": "{\"id\":15,\"title\":\"Review nftables rate limiting config\",\"status\":\"todo\",\"priority\":1,\"project_id\":5}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [{"type": "text", "text": "Done — task added to Pi Home Server, due by 7pm so you have time before the maintenance window."}]
        }
    ]',
    4,
    DATE_ADD(DATE_SUB(NOW(), INTERVAL 1 DAY), INTERVAL (7 * 60 + 45) MINUTE),
    DATE_ADD(DATE_SUB(NOW(), INTERVAL 1 DAY), INTERVAL (7 * 60 + 52) MINUTE)
);

INSERT INTO conversations (session_id, name, history, message_count, created_at, updated_at)
VALUES (
    'session-seed-b',
    'anything overdue?',
    '[
        {
            "role": "user",
            "content": [{"type": "text", "text": "Do I have anything overdue?"}]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_b1",
                    "name": "list_overdue_tasks",
                    "input": {}
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_b1",
                    "content": "{\"tasks\":[{\"id\":1,\"title\":\"Get three contractor quotes\",\"status\":\"done\",\"priority\":1,\"due_at\":\"2026-05-03T17:00:00Z\",\"project_id\":1}]}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [{"type": "text", "text": "One overdue item: Get three contractor quotes from the Home Renovation project, due 3 days ago. It is already marked done though, so nothing actually needs attention."}]
        },
        {
            "role": "user",
            "content": [{"type": "text", "text": "Good. Mark the iron shaft decision task as done too, I went with the X100."}]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_b2",
                    "name": "list_tasks",
                    "input": {"status": "in_progress", "project_id": 4}
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_b2",
                    "content": "{\"tasks\":[{\"id\":10,\"title\":\"Decide on iron shaft - X100 vs KBS Tour X\",\"status\":\"in_progress\",\"priority\":1,\"project_id\":4}]}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_b3",
                    "name": "update_task_status",
                    "input": {"id": 10, "status": "done"}
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_b3",
                    "content": "{\"id\":10,\"title\":\"Decide on iron shaft - X100 vs KBS Tour X\",\"status\":\"done\",\"priority\":1,\"project_id\":4}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [{"type": "text", "text": "Done — iron shaft task marked complete. Good call on the X100."}]
        },
        {
            "role": "user",
            "content": [{"type": "text", "text": "Remind me to order the shafts from Golfsmith every morning at 9 until I cancel it."}]
        },
        {
            "role": "assistant",
            "content": [
                {
                    "type": "tool_use",
                    "id": "tool_seed_b4",
                    "name": "create_reminder",
                    "input": {
                        "message": "Order X100 shafts from Golfsmith.",
                        "remind_at": "2026-05-07T09:00:00",
                        "recurrence_type": "daily",
                        "recurrence_rule": "{\"interval\": 1}"
                    }
                }
            ]
        },
        {
            "role": "user",
            "content": [
                {
                    "type": "tool_result",
                    "tool_use_id": "tool_seed_b4",
                    "content": "{\"id\":9,\"message\":\"Order X100 shafts from Golfsmith.\",\"remind_at\":\"2026-05-07T09:00:00Z\",\"recurrence_type\":\"daily\",\"is_active\":true}"
                }
            ]
        },
        {
            "role": "assistant",
            "content": [{"type": "text", "text": "Reminder set — you will get a daily nudge at 9am to order the shafts. It will keep firing until you delete it."}]
        }
    ]',
    5,
    DATE_ADD(DATE(NOW()), INTERVAL (8 * 60 + 30) MINUTE),
    DATE_ADD(DATE(NOW()), INTERVAL (8 * 60 + 58) MINUTE)
);
