package dal

import (
	"context"
	"fmt"
	"time"

	"fake.com/nilspcarlson/internal/models"
)

// ============================================================
// Events
// ============================================================

const eventColumns = `
	id, title, description, start_at, end_at, location, all_day, project_id, created_at, updated_at
`

func scanEvent(row interface {
	Scan(dest ...any) error
}) (*models.Event, error) {
	e := &models.Event{}
	var startAt, endAt, createdAt, updatedAt string
	var allDay int
	err := row.Scan(
		&e.ID,
		&e.Title,
		nullString{&e.Description},
		&startAt,
		&endAt,
		nullString{&e.Location},
		&allDay,
		nullInt64{&e.ProjectID},
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.AllDay = allDay == 1
	if e.StartAt, err = parseTime(startAt); err != nil {
		return nil, fmt.Errorf("event start_at: %w", err)
	}
	if e.EndAt, err = parseTime(endAt); err != nil {
		return nil, fmt.Errorf("event end_at: %w", err)
	}
	if e.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("event created_at: %w", err)
	}
	if e.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("event updated_at: %w", err)
	}
	return e, nil
}

// GetEvent returns a single event by ID.
func GetEvent(ctx context.Context, id int64) (*models.Event, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+eventColumns+`FROM events WHERE id = ?`, id)
	e, err := scanEvent(row)
	if err != nil {
		return nil, fmt.Errorf("GetEvent %d: %w", id, err)
	}
	return e, nil
}

// ListEventsInRange returns all events whose start_at falls within [from, to].
func ListEventsInRange(ctx context.Context, from, to time.Time) ([]*models.Event, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+eventColumns+`
		FROM   events
		WHERE  start_at >= ? AND start_at <= ?
		ORDER  BY start_at`,
		formatTime(from), formatTime(to),
	)
	if err != nil {
		return nil, fmt.Errorf("ListEventsInRange: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// ListEventsByProject returns all events for a given project.
func ListEventsByProject(ctx context.Context, projectID int64) ([]*models.Event, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+eventColumns+`
		FROM   events
		WHERE  project_id = ?
		ORDER  BY start_at`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListEventsByProject %d: %w", projectID, err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// CreateEvent inserts a new event and returns it with its assigned ID.
func CreateEvent(ctx context.Context, e *models.Event) (*models.Event, error) {
	allDay := 0
	if e.AllDay {
		allDay = 1
	}
	res, err := DB.ExecContext(ctx, `
		INSERT INTO events (title, description, start_at, end_at, location, all_day, project_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.Title,
		ptrStringVal(e.Description),
		formatTime(e.StartAt),
		formatTime(e.EndAt),
		ptrStringVal(e.Location),
		allDay,
		ptrInt64Val(e.ProjectID),
	)
	if err != nil {
		return nil, fmt.Errorf("CreateEvent: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetEvent(ctx, id)
}

// UpdateEvent updates mutable fields on an existing event.
func UpdateEvent(ctx context.Context, e *models.Event) (*models.Event, error) {
	allDay := 0
	if e.AllDay {
		allDay = 1
	}
	_, err := DB.ExecContext(ctx, `
		UPDATE events
		SET title = ?, description = ?, start_at = ?, end_at = ?,
		    location = ?, all_day = ?, project_id = ?
		WHERE id = ?`,
		e.Title,
		ptrStringVal(e.Description),
		formatTime(e.StartAt),
		formatTime(e.EndAt),
		ptrStringVal(e.Location),
		allDay,
		ptrInt64Val(e.ProjectID),
		e.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateEvent %d: %w", e.ID, err)
	}
	return GetEvent(ctx, e.ID)
}

// DeleteEvent deletes an event. Attached reminders will cascade delete.
func DeleteEvent(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM events WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteEvent %d: %w", id, err)
	}
	return nil
}

// scanEvents is a shared helper used by list functions.
func scanEvents(rows interface {
	Scan(dest ...any) error
	Next() bool
	Err() error
	Close() error
}) ([]*models.Event, error) {
	var events []*models.Event
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("scanEvents: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
