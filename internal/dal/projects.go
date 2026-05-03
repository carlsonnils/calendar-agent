package dal

import (
	"context"
	"fmt"

	"calendar/internal/models"
)

// ============================================================
// Projects
// ============================================================

const projectColumns = `
	id, name, description, status, color, due_date, created_at, updated_at
`

func scanProject(row interface {
	Scan(dest ...any) error
}) (*models.Project, error) {
	p := &models.Project{}
	var createdAt, updatedAt string
	err := row.Scan(
		&p.ID,
		&p.Name,
		nullString{&p.Description},
		&p.Status,
		nullString{&p.Color},
		nullTime{&p.DueDate},
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if p.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("project created_at: %w", err)
	}
	if p.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("project updated_at: %w", err)
	}
	return p, nil
}

// GetProject returns a single project by ID.
func GetProject(ctx context.Context, id int64) (*models.Project, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+projectColumns+`FROM projects WHERE id = ?`, id)
	p, err := scanProject(row)
	if err != nil {
		return nil, fmt.Errorf("GetProject %d: %w", id, err)
	}
	return p, nil
}

// ListProjects returns all projects, optionally filtered by status.
// Pass an empty string to return all statuses.
func ListProjects(ctx context.Context, status models.ProjectStatus) ([]*models.Project, error) {
	var rows interface {
		Scan(dest ...any) error
		Next() bool
		Err() error
		Close() error
	}

	var err error
	if status == "" {
		r, e := DB.QueryContext(ctx,
			`SELECT`+projectColumns+`FROM projects ORDER BY name`)
		rows, err = r, e
	} else {
		r, e := DB.QueryContext(ctx,
			`SELECT`+projectColumns+`FROM projects WHERE status = ? ORDER BY name`, status)
		rows, err = r, e
	}
	if err != nil {
		return nil, fmt.Errorf("ListProjects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("ListProjects scan: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// CreateProject inserts a new project and returns it with its assigned ID.
func CreateProject(ctx context.Context, p *models.Project) (*models.Project, error) {
	if p.Status == "" {
		p.Status = models.ProjectStatusActive
	}
	res, err := DB.ExecContext(ctx, `
		INSERT INTO projects (name, description, status, color, due_date)
		VALUES (?, ?, ?, ?, ?)`,
		p.Name,
		ptrStringVal(p.Description),
		p.Status,
		ptrStringVal(p.Color),
		ptrDateVal(p.DueDate),
	)
	if err != nil {
		return nil, fmt.Errorf("CreateProject: %w", err)
	}
	id, _ := res.LastInsertId()
	return GetProject(ctx, id)
}

// UpdateProject updates mutable fields on an existing project.
func UpdateProject(ctx context.Context, p *models.Project) (*models.Project, error) {
	_, err := DB.ExecContext(ctx, `
		UPDATE projects
		SET name = ?, description = ?, status = ?, color = ?, due_date = ?
		WHERE id = ?`,
		p.Name,
		ptrStringVal(p.Description),
		p.Status,
		ptrStringVal(p.Color),
		ptrDateVal(p.DueDate),
		p.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateProject %d: %w", p.ID, err)
	}
	return GetProject(ctx, p.ID)
}

// DeleteProject deletes a project. Tasks and events with this project_id
// will have their project_id set to NULL (ON DELETE SET NULL).
func DeleteProject(ctx context.Context, id int64) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("DeleteProject %d: %w", id, err)
	}
	return nil
}
