// Package models defines the data types that map directly to the database schema.
// All datetime fields are stored as ISO8601 strings in SQLite and parsed to
// time.Time at the boundary. NULLable columns use pointer types.
package models

import (
	"encoding/json"
	"time"
)

// ============================================================
// Project
// ============================================================

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusPaused    ProjectStatus = "paused"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusArchived  ProjectStatus = "archived"
)

type Project struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description"`
	Status      ProjectStatus `json:"status"`
	Color       *string       `json:"color"`
	DueDate     *time.Time    `json:"due_date"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// ============================================================
// Event
// ============================================================

type Event struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Location    *string   `json:"location"`
	AllDay      bool      `json:"all_day"`
	ProjectID   *int64    `json:"project_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================
// Task
// ============================================================

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskPriority int

const (
	PriorityHigh   TaskPriority = 1
	PriorityMedium TaskPriority = 2
	PriorityLow    TaskPriority = 3
)

type Task struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	DueAt       *time.Time   `json:"due_at"`
	ProjectID   *int64       `json:"project_id"`
	ParentID    *int64       `json:"parent_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// Populated by GetTaskWithSubtasks — not stored in DB
	Subtasks []Task `json:"subtasks,omitempty"`
}

// ============================================================
// Reminder
// ============================================================

type RecurrenceType string

const (
	RecurrenceMinutely RecurrenceType = "minutely"
	RecurrenceHourly   RecurrenceType = "hourly"
	RecurrenceDaily    RecurrenceType = "daily"
	RecurrenceWeekly   RecurrenceType = "weekly"
	RecurrenceMonthly  RecurrenceType = "monthly"
	RecurrenceYearly   RecurrenceType = "yearly"
	RecurrenceCustom   RecurrenceType = "custom"
)

type Reminder struct {
	ID             int64           `json:"id"`
	Message        string          `json:"message"`
	RemindAt       time.Time       `json:"remind_at"`
	EventID        *int64          `json:"event_id"`
	TaskID         *int64          `json:"task_id"`
	RecurrenceType *RecurrenceType `json:"recurrence_type"`
	RecurrenceRule *string         `json:"recurrence_rule"` // raw JSON string
	RecurrenceEnd  *time.Time      `json:"recurrence_end"`
	IsActive       bool            `json:"is_active"`
	LastFiredAt    *time.Time      `json:"last_fired_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// ============================================================
// LogEntry
// ============================================================

type EntryType string

const (
	EntryTypeNote      EntryType = "note"
	EntryTypeDiary     EntryType = "diary"
	EntryTypeWorkLog   EntryType = "work_log"
	EntryTypeThought   EntryType = "thought"
	EntryTypeMood      EntryType = "mood"
	EntryTypeIdea      EntryType = "idea"
	EntryTypeWin       EntryType = "win"
	EntryTypeGratitude EntryType = "gratitude"
)

type LogEntry struct {
	ID               int64     `json:"id"`
	EntryType        EntryType `json:"entry_type"`
	Title            *string   `json:"title"`
	Body             string    `json:"body"`
	MoodScore        *int      `json:"mood_score"`
	Situation        *string   `json:"situation"`
	AutomaticThought *string   `json:"automatic_thought"`
	Reframe          *string   `json:"reframe"`
	ProjectID        *int64    `json:"project_id"`
	TaskID           *int64    `json:"task_id"`
	Tags             []string  `json:"tags"` // stored as JSON array in DB
	EntryDate        time.Time `json:"entry_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ============================================================
// Agenda — cross-cutting read model for get_agenda queries
// ============================================================

// AgendaItem is a unified view of anything happening in a time window.
// Used by the agent's get_agenda tool to answer "what's on my plate".
type AgendaItem struct {
	Kind      string    `json:"kind"` // "event" | "task" | "reminder"
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	At        time.Time `json:"at"` // start_at / due_at / remind_at
	ProjectID *int64    `json:"project_id"`
	Extra     string    `json:"extra"` // status for tasks, location for events
}

// ============================================================
// Conversation
// ============================================================

//
// The History field is stored as a JSON blob — the same []Message slice the
// agent package uses. The dal package doesn't import agent to avoid a cycle,
// so it works with []byte / json.RawMessage at the boundary.

type Conversation struct {
	SessionID    string 		 `json:"session_id"`
	Name    	 string			 `json:"name"`
	History      json.RawMessage `json:"history"` // []agent.Message serialised
	MessageCount int			 `json:"message_count"`
	CreatedAt    time.Time		 `json:"created_at"`
	UpdatedAt    time.Time		 `json:"updated_at"`
}
