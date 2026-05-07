// Package dal (Data Access Layer) provides typed functions for every
// database operation. All SQL lives here — nothing above this layer
// touches the database directly.
//
// Time handling:
//
//	SQLite stores datetimes as TEXT in ISO8601 format ("2006-01-02T15:04:05"
//	or "2006-01-02 15:04:05"). The helpers below normalise both formats on
//	scan so callers always receive time.Time values in UTC.
package dal

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DB is set by the database package after Open() succeeds.
// All DAL functions use this handle.
var DB *sql.DB

// ============================================================
// Time helpers
// ============================================================

// sqliteTimeFormats lists the layouts SQLite datetime() can produce.
var sqliteTimeFormats = []string{
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05Z",
	"2006-01-02",
}

// parseTime parses an ISO8601 string from SQLite into a time.Time (UTC).
func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range sqliteTimeFormats {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parseTime: unrecognised format %q", s)
}

// formatTime converts a time.Time to the ISO8601 string SQLite expects.
func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05")
}

// formatDate converts a time.Time to a date-only string for date columns.
func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// ============================================================
// Nullable scan targets
// These types implement sql.Scanner and convert between SQLite
// TEXT/NULL and Go pointer types, handling time parsing along the way.
// ============================================================

// nullTime scans a NULLable TEXT datetime column into *time.Time.
type nullTime struct{ ptr **time.Time }

func (n nullTime) Scan(src any) error {
	if src == nil {
		*n.ptr = nil
		return nil
	}
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("nullTime.Scan: expected string, got %T", src)
	}
	t, err := parseTime(s)
	if err != nil {
		return err
	}
	*n.ptr = &t
	return nil
}

// nullString scans a NULLable TEXT column into *string.
type nullString struct{ ptr **string }

func (n nullString) Scan(src any) error {
	if src == nil {
		*n.ptr = nil
		return nil
	}
	switch v := src.(type) {
	case string:
		s := v
		*n.ptr = &s
	case []byte:
		s := string(v)
		*n.ptr = &s
	default:
		return fmt.Errorf("nullString.Scan: expected string, got %T", src)
	}
	return nil
}

// nullInt64 scans a NULLable INTEGER column into *int64.
type nullInt64 struct{ ptr **int64 }

func (n nullInt64) Scan(src any) error {
	if src == nil {
		*n.ptr = nil
		return nil
	}
	switch v := src.(type) {
	case int64:
		i := v
		*n.ptr = &i
	case float64:
		i := int64(v)
		*n.ptr = &i
	default:
		return fmt.Errorf("nullInt64.Scan: expected int64, got %T", src)
	}
	return nil
}

// nullInt scans a NULLable INTEGER column into *int.
type nullInt struct{ ptr **int }

func (n nullInt) Scan(src any) error {
	if src == nil {
		*n.ptr = nil
		return nil
	}
	switch v := src.(type) {
	case int64:
		i := int(v)
		*n.ptr = &i
	case float64:
		i := int(v)
		*n.ptr = &i
	default:
		return fmt.Errorf("nullInt.Scan: expected int64, got %T", src)
	}
	return nil
}

// jsonTags scans the JSON array stored in the tags column into []string.
type jsonTags struct{ ptr *[]string }

func (j jsonTags) Scan(src any) error {
	*j.ptr = []string{}
	if src == nil {
		return nil
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("jsonTags.Scan: expected string, got %T", src)
	}
	return json.Unmarshal([]byte(raw), j.ptr)
}

// rawHistory scans the JSON array stored in the history column into json.RawMessage
type rawHistory struct{ ptr *json.RawMessage }

func (h rawHistory) Scan(src any) error {
    if src == nil {
        *h.ptr = nil
        return nil
    }
    switch v := src.(type) {
    case string:
        *h.ptr = json.RawMessage(v)
    case []byte:
        *h.ptr = json.RawMessage(v) // safe: database/sql won't reuse this slice
    default:
        return fmt.Errorf("rawHistory.Scan: expected string or []byte, got %T", src)
    }
    return nil
}

// ============================================================
// Pointer helpers for building INSERT/UPDATE args
// ============================================================

func ptrStringVal(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func ptrInt64Val(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func ptrTimeVal(p *time.Time) any {
	if p == nil {
		return nil
	}
	return formatTime(*p)
}

func ptrDateVal(p *time.Time) any {
	if p == nil {
		return nil
	}
	return formatDate(*p)
}

// marshalTags encodes a []string into a JSON array string for storage.
func marshalTags(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

// ============================================================
// Shared query helper
// ============================================================

// execContext is a minimal interface satisfied by both *sql.DB and *sql.Tx.
type execContext interface {
	ExecContext(ctx interface {
		Deadline() (time.Time, bool)
		Done() <-chan struct{}
		Err() error
		Value(any) any
	}, query string, args ...any) (sql.Result, error)
}
