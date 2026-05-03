package dal

import (
	"context"
	"fmt"

	"calendar/internal/models"
)

// ============================================================
// Tasks
// ============================================================

const taskColumns = `
	id, title, description, status, priority, due_at, project_id, parent_id, created_at, updated_at
`

func scanTask(row interface {
	Scan(dest ...any) error
}) (*models.Task, error) {
	t := &models.Task{}
	var createdAt, updatedAt string
	var priority int64
	var status string
	err := row.Scan(
		&t.ID,
		&t.Title,
		nullString{&t.Description},
		&status,
		&priority,
		nullTime{&t.DueAt},
		nullInt64{&t.ProjectID},
		nullInt64{&t.ParentID},
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	t.Status = models.TaskStatus(status)
	t.Priority = models.TaskPriority(priority)
	if t.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("task created_at: %w", err)
	}
	if t.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("task updated_at: %w", err)
	}
	return t, nil
}

// GetTask returns a single task by ID.
func GetTask(ctx context.Context, id int64) (*models.Task, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+taskColumns+`FROM tasks WHERE id = ?`, id)
	t, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("GetTask %d: %w", id, err)
	}
	return t, nil
}

// GetTaskWithSubtasks returns a task and populates its Subtasks field.
func GetTaskWithSubtasks(ctx context.Context, id int64) (*models.Task, error) {
	t, err := GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := DB.QueryContext(ctx,
		`SELECT`+taskColumns+`FROM tasks WHERE parent_id = ? ORDER BY priority, due_at`, id)
	if err != nil {
		return nil, fmt.Errorf("GetTaskWithSubtasks subtasks: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		sub, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		t.Subtasks = append(t.Subtasks, *sub)
	}
	return t, rows.Err()
}

// ListTasks returns tasks filtered by optional status and/or project.
// Pass empty string / nil to skip a filter.
func ListTasks(ctx context.Context, status models.TaskStatus, projectID *int64) ([]*models.Task, error) {
	query := `SELECT` + taskColumns + `FROM tasks WHERE parent_id IS NULL`
	args := []any{}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, string(status))
	}
	if projectID != nil {
		query += ` AND project_id = ?`
		args = append(args, *projectID)
	}
	query += ` ORDER BY priority, due_at`

	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListTasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("ListTasks scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListOverdueTasks returns all active tasks with a due_at in the past.
func ListOverdueTasks(ctx context.Context) ([]*models.Task, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT`+taskColumns+`
		FROM   tasks
		WHERE  due_at < datetime('now')
		  AND  status NOT IN ('done', 'cancelled')
		ORDER  BY due_at`,
	)
	if err != nil {
		return nil, fmt.Errorf("ListOverdueTasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// CreateTask inserts a new task and returns it with its assigned ID.
func CreateTask(ctx context.Context, t *models.Task) (*models.Task, error) {
	if t.Status == "" {
		t.Status = models.TaskStatusTodo
	}
	if t.Priority == 0 {
		t.Priority = models.PriorityMedium
	}
	res, err := DB.ExecContext(ctx, `
		INSERT INTO tasks (title, description, status, priority, due_at, project_id, parent_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.Title,
		ptrStringVal(t.Description),
		t.Status,
		t.Priority,
		ptrTimeVal(t.DueAt),
		ptrInt64Val(t.ProjectID),
		ptrInt64Val(t.ParentID),
	)
	if err != nil {
		return nil, fmt.Errorf("CreateTask: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetTask(ctx, id)
}

// UpdateTask updates mutable fields on an existing task.
func UpdateTask(ctx context.Context, t *models.Task) (*models.Task, error) {
	_, err := DB.ExecContext(ctx, `
		UPDATE tasks
		SET title = ?, description = ?, status = ?, priority = ?,
		    due_at = ?, project_id = ?, parent_id = ?
		WHERE id = ?`,
		t.Title,
		ptrStringVal(t.Description),
		t.Status,
		t.Priority,
		ptrTimeVal(t.DueAt),
		ptrInt64Val(t.ProjectID),
		ptrInt64Val(t.ParentID),
		t.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateTask %d: %w", t.ID, err)
	}
	return GetTask(ctx, t.ID)
}

// UpdateTaskStatus is a convenience helper for the common agent action
// of changing only the status field (e.g. "mark the report task as done").
func UpdateTaskStatus(ctx context.Context, id int64, status models.TaskStatus) (*models.Task, error) {
	_, err := DB.ExecContext(ctx,
		`UPDATE tasks SET status = ? WHERE id = ?`, string(status), id)
	if err != nil {
		return nil, fmt.Errorf("UpdateTaskStatus %d: %w", id, err)
	}
	return GetTask(ctx, id)
}

// DeleteTask deletes a task. Subtasks cascade delete due to ON DELETE CASCADE.
func DeleteTask(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteTask %d: %w", id, err)
	}
	return nil
}
