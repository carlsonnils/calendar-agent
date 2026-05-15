package dal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"fake.com/nilspcarlson/internal/models"
)

// conversations.go persists agent conversation history to SQLite so sessions
// survive server restarts and page reloads.

// helper function to scan a conversation rows
func scanConversation(row interface {
	Scan(dest ...any) error
}) (*models.Conversation, error) {
	var rec models.Conversation
	var createdAt, updatedAt string
	err := row.Scan(
		&rec.SessionID,
		&rec.Name,
		rawHistory{&rec.History},
		&rec.MessageCount,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("LoadConversation %q-%v: %w", rec.SessionID, rec.Name, err)
	}
	if rec.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, err
	}
	if rec.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, err
	}
	return &rec, nil
}

// LoadConversation returns the conversation record for sessionID, or nil if
// the session does not exist yet.
func ListConversations(ctx context.Context) ([]*models.Conversation, error) {
	rows, err := DB.QueryContext(ctx, `
		SELECT session_id, name, history, message_count, created_at, updated_at
		FROM   conversations
		ORDER BY updated_at DESC
		LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("ListConversations query: %w", err)
	}
	
	var conversations []*models.Conversation
	for rows.Next() {
		convo, err := scanConversation(rows)
		if err != nil {
			return conversations, fmt.Errorf("ListConversations scan: %w", err)
		}
		conversations = append(conversations, convo)
	}

	return conversations, nil
}

// LoadConversations returns the conversation record for sessionID, or nil if
// the session does not exist yet.
func LoadConversation(ctx context.Context, sessionID string) (*models.Conversation, error) {
	row := DB.QueryRowContext(ctx, `
		SELECT session_id, name, history, message_count, created_at, updated_at
		FROM   conversations
		WHERE  session_id = ?`, sessionID)

	rec, err := scanConversation(row)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

// SaveConversation upserts the conversation history for sessionID.
// messageCount should be the total number of user turns so far.
func SaveConversation(ctx context.Context, sessionID, name string, history json.RawMessage, messageCount int) error {
	histStr := string(history)
	_, err := DB.ExecContext(ctx, `
		INSERT INTO conversations (session_id, name, history, message_count)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET
			name          = excluded.name,
			history       = excluded.history,
			message_count = excluded.message_count`,
		sessionID, histStr, messageCount,
	)
	if err != nil {
		return fmt.Errorf("SaveConversation %q: %w", sessionID, err)
	}
	return nil
}

// DeleteConversation removes a session entirely (used by the "new chat" button).
func DeleteConversation(ctx context.Context, sessionID string) error {
	_, err := DB.ExecContext(ctx,
		`DELETE FROM conversations WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("DeleteConversation %q: %w", sessionID, err)
	}
	return nil
}

// PruneOldConversations deletes sessions that have not been updated in
// olderThan duration. Call periodically (e.g. daily) to keep the DB lean.
func PruneOldConversations(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan).Format("2006-01-02T15:04:05")
	res, err := DB.ExecContext(ctx,
		`DELETE FROM conversations WHERE updated_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("PruneOldConversations: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
