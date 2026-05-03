package dal

import (
	"context"
	"fmt"
	"time"

	"calendar/internal/models"
)

// ============================================================
// Log Entries
// ============================================================

const logEntryColumns = `
	id, entry_type, title, body,
	mood_score, situation, automatic_thought, reframe,
	project_id, task_id, tags, entry_date,
	created_at, updated_at
`

func scanLogEntry(row interface {
	Scan(dest ...any) error
}) (*models.LogEntry, error) {
	e := &models.LogEntry{}
	var entryType, body, entryDate, createdAt, updatedAt string
	err := row.Scan(
		&e.ID,
		&entryType,
		nullString{&e.Title},
		&body,
		nullInt{&e.MoodScore},
		nullString{&e.Situation},
		nullString{&e.AutomaticThought},
		nullString{&e.Reframe},
		nullInt64{&e.ProjectID},
		nullInt64{&e.TaskID},
		jsonTags{&e.Tags},
		&entryDate,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	e.EntryType = models.EntryType(entryType)
	e.Body = body
	if e.EntryDate, err = parseTime(entryDate); err != nil {
		return nil, fmt.Errorf("log entry_date: %w", err)
	}
	if e.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("log created_at: %w", err)
	}
	if e.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("log updated_at: %w", err)
	}
	return e, nil
}

// GetLogEntry returns a single log entry by ID.
func GetLogEntry(ctx context.Context, id int64) (*models.LogEntry, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+logEntryColumns+`FROM log_entries WHERE id = ?`, id)
	e, err := scanLogEntry(row)
	if err != nil {
		return nil, fmt.Errorf("GetLogEntry %d: %w", id, err)
	}
	return e, nil
}

// ListLogEntries returns entries filtered by optional entry type and date range.
// Pass empty entryType to return all types. Pass zero times to skip date filter.
func ListLogEntries(ctx context.Context, entryType models.EntryType, from, to time.Time) ([]*models.LogEntry, error) {
	query := `SELECT` + logEntryColumns + `FROM log_entries WHERE 1=1`
	args := []any{}

	if entryType != "" {
		query += ` AND entry_type = ?`
		args = append(args, string(entryType))
	}
	if !from.IsZero() {
		query += ` AND entry_date >= ?`
		args = append(args, formatTime(from))
	}
	if !to.IsZero() {
		query += ` AND entry_date <= ?`
		args = append(args, formatTime(to))
	}
	query += ` ORDER BY entry_date DESC`

	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListLogEntries: %w", err)
	}
	defer rows.Close()

	var entries []*models.LogEntry
	for rows.Next() {
		e, err := scanLogEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("ListLogEntries scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListLogEntriesByTag returns entries that contain a specific tag.
// Uses SQLite's json_each() to query the JSON tags array.
func ListLogEntriesByTag(ctx context.Context, tag string) ([]*models.LogEntry, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT `+logEntryColumns+`
		FROM   log_entries, json_each(log_entries.tags)
		WHERE  json_each.value = ?
		ORDER  BY entry_date DESC`,
		tag,
	)
	if err != nil {
		return nil, fmt.Errorf("ListLogEntriesByTag %q: %w", tag, err)
	}
	defer rows.Close()

	var entries []*models.LogEntry
	for rows.Next() {
		e, err := scanLogEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListLogEntriesByProject returns entries linked to a project.
func ListLogEntriesByProject(ctx context.Context, projectID int64) ([]*models.LogEntry, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+logEntryColumns+`
		FROM   log_entries
		WHERE  project_id = ?
		ORDER  BY entry_date DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("ListLogEntriesByProject %d: %w", projectID, err)
	}
	defer rows.Close()

	var entries []*models.LogEntry
	for rows.Next() {
		e, err := scanLogEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// MoodTrend returns (entry_date, mood_score) pairs for entries that have a
// mood score, ordered by date. Used by the agent for "how has my mood been".
type MoodPoint struct {
	Date  time.Time
	Score int
}

func MoodTrend(ctx context.Context, from, to time.Time) ([]MoodPoint, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT entry_date, mood_score
		FROM   log_entries
		WHERE  mood_score IS NOT NULL
		  AND  entry_date >= ?
		  AND  entry_date <= ?
		ORDER  BY entry_date`,
		formatTime(from), formatTime(to),
	)
	if err != nil {
		return nil, fmt.Errorf("MoodTrend: %w", err)
	}
	defer rows.Close()

	var points []MoodPoint
	for rows.Next() {
		var dateStr string
		var score int
		if err := rows.Scan(&dateStr, &score); err != nil {
			return nil, err
		}
		t, err := parseTime(dateStr)
		if err != nil {
			return nil, err
		}
		points = append(points, MoodPoint{Date: t, Score: score})
	}
	return points, rows.Err()
}

// CreateLogEntry inserts a new log entry and returns it with its assigned ID.
func CreateLogEntry(ctx context.Context, e *models.LogEntry) (*models.LogEntry, error) {
	if e.EntryType == "" {
		e.EntryType = models.EntryTypeNote
	}
	entryDate := e.EntryDate
	if entryDate.IsZero() {
		entryDate = time.Now().UTC()
	}

	res, err := DB.ExecContext(ctx, `
		INSERT INTO log_entries
		    (entry_type, title, body,
		     mood_score, situation, automatic_thought, reframe,
		     project_id, task_id, tags, entry_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.EntryType,
		ptrStringVal(e.Title),
		e.Body,
		func() any {
			if e.MoodScore == nil {
				return nil
			}
			return *e.MoodScore
		}(),
		ptrStringVal(e.Situation),
		ptrStringVal(e.AutomaticThought),
		ptrStringVal(e.Reframe),
		ptrInt64Val(e.ProjectID),
		ptrInt64Val(e.TaskID),
		marshalTags(e.Tags),
		formatTime(entryDate),
	)
	if err != nil {
		return nil, fmt.Errorf("CreateLogEntry: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetLogEntry(ctx, id)
}

// UpdateLogEntry updates mutable fields on an existing log entry.
func UpdateLogEntry(ctx context.Context, e *models.LogEntry) (*models.LogEntry, error) {
	_, err := DB.ExecContext(ctx, `
		UPDATE log_entries
		SET entry_type = ?, title = ?, body = ?,
		    mood_score = ?, situation = ?, automatic_thought = ?, reframe = ?,
		    project_id = ?, task_id = ?, tags = ?, entry_date = ?
		WHERE id = ?`,
		e.EntryType,
		ptrStringVal(e.Title),
		e.Body,
		func() any {
			if e.MoodScore == nil {
				return nil
			}
			return *e.MoodScore
		}(),
		ptrStringVal(e.Situation),
		ptrStringVal(e.AutomaticThought),
		ptrStringVal(e.Reframe),
		ptrInt64Val(e.ProjectID),
		ptrInt64Val(e.TaskID),
		marshalTags(e.Tags),
		formatTime(e.EntryDate),
		e.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateLogEntry %d: %w", e.ID, err)
	}
	return GetLogEntry(ctx, e.ID)
}

// DeleteLogEntry deletes a log entry.
func DeleteLogEntry(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM log_entries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteLogEntry %d: %w", id, err)
	}
	return nil
}
