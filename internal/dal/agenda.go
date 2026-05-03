package dal

import (
	"context"
	"fmt"
	"time"

	"calendar/internal/models"
)

// ============================================================
// Agenda — unified cross-table read for the agent's get_agenda tool
// ============================================================

// GetAgenda returns a unified, time-ordered list of everything happening
// within [from, to]: events that start in the window, tasks due in the
// window, and reminders firing in the window.
//
// This is the primary read path for agent queries like:
//
//	"what's on my plate today?"
//	"what do I have coming up this week?"
//	"show me everything for the next 3 days"
func GetAgenda(ctx context.Context, from, to time.Time) ([]models.AgendaItem, error) {
	// Single UNION query — one round trip, results arrive pre-sorted by time.
	const query = `
		SELECT 'event'    AS kind,
		       e.id,
		       e.title,
		       e.start_at AS at,
		       e.project_id,
		       COALESCE(e.location, '') AS extra
		FROM   events e
		WHERE  e.start_at >= ? AND e.start_at <= ?

		UNION ALL

		SELECT 'task'     AS kind,
		       t.id,
		       t.title,
		       t.due_at   AS at,
		       t.project_id,
		       t.status   AS extra
		FROM   tasks t
		WHERE  t.due_at >= ? AND t.due_at <= ?
		  AND  t.status NOT IN ('done', 'cancelled')

		UNION ALL

		SELECT 'reminder' AS kind,
		       r.id,
		       r.message  AS title,
		       r.remind_at AS at,
		       NULL       AS project_id,
		       COALESCE(r.recurrence_type, 'one-shot') AS extra
		FROM   reminders r
		WHERE  r.remind_at >= ? AND r.remind_at <= ?
		  AND  r.is_active = 1

		ORDER  BY at
	`

	f := formatTime(from)
	t2 := formatTime(to)

	rows, err := DB.QueryContext(ctx, query, f, t2, f, t2, f, t2)
	if err != nil {
		return nil, fmt.Errorf("GetAgenda: %w", err)
	}
	defer rows.Close()

	var items []models.AgendaItem
	for rows.Next() {
		var item models.AgendaItem
		var atStr string
		err := rows.Scan(
			&item.Kind,
			&item.ID,
			&item.Title,
			&atStr,
			nullInt64{&item.ProjectID},
			&item.Extra,
		)
		if err != nil {
			return nil, fmt.Errorf("GetAgenda scan: %w", err)
		}
		item.At, err = parseTime(atStr)
		if err != nil {
			return nil, fmt.Errorf("GetAgenda parse time: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// TodayAgenda is a convenience wrapper that calls GetAgenda for today.
func TodayAgenda(ctx context.Context) ([]models.AgendaItem, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)
	return GetAgenda(ctx, startOfDay, endOfDay)
}

// WeekAgenda is a convenience wrapper that calls GetAgenda for the next 7 days.
func WeekAgenda(ctx context.Context) ([]models.AgendaItem, error) {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return GetAgenda(ctx, startOfDay, startOfDay.Add(7*24*time.Hour))
}
