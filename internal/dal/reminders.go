package dal

import (
	"context"
	"fmt"
	"time"

	"fake.com/nilspcarlson/internal/models"
)

// ============================================================
// Reminders
// ============================================================

const reminderColumns = `
	id, message, remind_at,
	event_id, task_id,
	recurrence_type, recurrence_rule, recurrence_end,
	is_active, last_fired_at,
	created_at, updated_at
`

func scanReminder(row interface {
	Scan(dest ...any) error
}) (*models.Reminder, error) {
	r := &models.Reminder{}
	var remindAt, createdAt, updatedAt string
	var isActive int
	var recurrenceType *string

	err := row.Scan(
		&r.ID,
		&r.Message,
		&remindAt,
		nullInt64{&r.EventID},
		nullInt64{&r.TaskID},
		nullString{&recurrenceType},
		nullString{&r.RecurrenceRule},
		nullTime{&r.RecurrenceEnd},
		&isActive,
		nullTime{&r.LastFiredAt},
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	r.IsActive = isActive == 1

	if recurrenceType != nil {
		rt := models.RecurrenceType(*recurrenceType)
		r.RecurrenceType = &rt
	}

	if r.RemindAt, err = parseTime(remindAt); err != nil {
		return nil, fmt.Errorf("reminder remind_at: %w", err)
	}
	if r.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("reminder created_at: %w", err)
	}
	if r.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("reminder updated_at: %w", err)
	}
	return r, nil
}

// GetReminder returns a single reminder by ID.
func GetReminder(ctx context.Context, id int64) (*models.Reminder, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+reminderColumns+`FROM reminders WHERE id = ?`, id)
	r, err := scanReminder(row)
	if err != nil {
		return nil, fmt.Errorf("GetReminder %d: %w", id, err)
	}
	return r, nil
}

// ListActiveReminders returns all reminders that are active.
func ListActiveReminders(ctx context.Context) ([]*models.Reminder, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+reminderColumns+`
		FROM   reminders
		WHERE  is_active = 1
		ORDER  BY remind_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("ListActiveReminders: %w", err)
	}
	defer rows.Close()
	return scanReminders(rows)
}

// ListRemindersByEvent returns all reminders attached to an event.
func ListRemindersByEvent(ctx context.Context, eventID int64) ([]*models.Reminder, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+reminderColumns+`
		FROM   reminders
		WHERE  event_id = ?
		ORDER  BY remind_at`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListRemindersByEvent %d: %w", eventID, err)
	}
	defer rows.Close()
	return scanReminders(rows)
}

// ListRemindersByTask returns all reminders attached to a task.
func ListRemindersByTask(ctx context.Context, taskID int64) ([]*models.Reminder, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+reminderColumns+`
		FROM   reminders
		WHERE  task_id = ?
		ORDER  BY remind_at`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListRemindersByTask %d: %w", taskID, err)
	}
	defer rows.Close()
	return scanReminders(rows)
}

// DueReminders returns all active reminders whose remind_at <= now.
// This is the query the daemon calls every minute.
func DueReminders(ctx context.Context) ([]*models.Reminder, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+reminderColumns+`
		FROM   reminders
		WHERE  is_active = 1
		  AND  remind_at <= datetime('now')
		ORDER  BY remind_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("DueReminders: %w", err)
	}
	defer rows.Close()
	return scanReminders(rows)
}

// CreateReminder inserts a new reminder and returns it with its assigned ID.
func CreateReminder(ctx context.Context, r *models.Reminder) (*models.Reminder, error) {
	var recurrenceType any
	if r.RecurrenceType != nil {
		recurrenceType = string(*r.RecurrenceType)
	}

	res, err := DB.ExecContext(ctx, `
		INSERT INTO reminders
		    (message, remind_at, event_id, task_id,
		     recurrence_type, recurrence_rule, recurrence_end, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Message,
		formatTime(r.RemindAt),
		ptrInt64Val(r.EventID),
		ptrInt64Val(r.TaskID),
		recurrenceType,
		ptrStringVal(r.RecurrenceRule),
		ptrDateVal(r.RecurrenceEnd),
		1, // always start active
	)
	if err != nil {
		return nil, fmt.Errorf("CreateReminder: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetReminder(ctx, id)
}

// MarkFired marks a one-shot reminder as inactive after it fires.
func MarkFired(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `
		UPDATE reminders
		SET is_active = 0, last_fired_at = datetime('now')
		WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("MarkFired %d: %w", id, err)
	}
	return nil
}

// AdvanceReminder updates remind_at to the next fire time for a recurring
// reminder after it fires, and records last_fired_at.
// The caller is responsible for computing nextFireAt from the recurrence rule.
func AdvanceReminder(ctx context.Context, id int64, nextFireAt time.Time) error {
	_, err := DB.ExecContext(ctx, `
		UPDATE reminders
		SET remind_at = ?, last_fired_at = datetime('now')
		WHERE id = ?`,
		formatTime(nextFireAt), id)
	if err != nil {
		return fmt.Errorf("AdvanceReminder %d: %w", id, err)
	}
	return nil
}

// DeactivateReminder sets is_active = 0 without marking as fired.
// Used when a recurring reminder's recurrence_end has passed.
func DeactivateReminder(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx,
		`UPDATE reminders SET is_active = 0 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeactivateReminder %d: %w", id, err)
	}
	return nil
}

// DeleteReminder deletes a reminder entirely.
func DeleteReminder(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM reminders WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteReminder %d: %w", id, err)
	}
	return nil
}

func scanReminders(rows interface {
	Scan(dest ...any) error
	Next() bool
	Err() error
	Close() error
}) ([]*models.Reminder, error) {
	var reminders []*models.Reminder
	for rows.Next() {
		r, err := scanReminder(rows)
		if err != nil {
			return nil, fmt.Errorf("scanReminders: %w", err)
		}
		reminders = append(reminders, r)
	}
	return reminders, rows.Err()
}
