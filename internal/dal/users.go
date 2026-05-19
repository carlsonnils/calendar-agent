package dal

import (
	"context"
	"fmt"
	"time"
)

// ============================================================
// Users
// ============================================================

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const userColumns = `
	id, name, username, password, created_at, updated_at
`

func scanUser(row interface {
	Scan(dest ...any) error
}) (*User, error) {
	u := &User{}
	var createdAt, updatedAt string
	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Username,
		&u.Password,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}
	if u.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("event created_at: %w", err)
	}
	if u.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("event updated_at: %w", err)
	}
	return u, nil
}

// GetUser returns a single user by username
func GetUser(ctx context.Context, username string) (*User, error) {
	row := DB.QueryRowContext(ctx,
		`SELECT`+userColumns+`FROM users WHERE username = ?`, username)
	u, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("dal.GetUser %s: %w", username, err)
	}
	return u, nil
}
