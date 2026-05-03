package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"calendar/internal/dal"
	"calendar/internal/models"
)

// handlers.go maps every tool name to a function that:
//   1. Unmarshals the raw JSON input from the API response
//   2. Calls the appropriate DAL function(s)
//   3. Returns a JSON string to inject back as the tool_result

// DispatchTool routes a tool_use block to the correct handler.
// Returns a JSON string that becomes the tool_result content.
func DispatchTool(ctx context.Context, name string, inputJSON json.RawMessage) (string, error) {
	switch name {
	// Agenda
	case "get_agenda":
		return handleGetAgenda(ctx, inputJSON)

	// Projects
	case "list_projects":
		return handleListProjects(ctx, inputJSON)
	case "create_project":
		return handleCreateProject(ctx, inputJSON)
	case "update_project":
		return handleUpdateProject(ctx, inputJSON)
	case "delete_project":
		return handleDeleteProject(ctx, inputJSON)

	// Events
	case "list_events":
		return handleListEvents(ctx, inputJSON)
	case "create_event":
		return handleCreateEvent(ctx, inputJSON)
	case "update_event":
		return handleUpdateEvent(ctx, inputJSON)
	case "delete_event":
		return handleDeleteEvent(ctx, inputJSON)

	// Tasks
	case "list_tasks":
		return handleListTasks(ctx, inputJSON)
	case "get_task":
		return handleGetTask(ctx, inputJSON)
	case "list_overdue_tasks":
		return handleListOverdueTasks(ctx, inputJSON)
	case "create_task":
		return handleCreateTask(ctx, inputJSON)
	case "update_task":
		return handleUpdateTask(ctx, inputJSON)
	case "update_task_status":
		return handleUpdateTaskStatus(ctx, inputJSON)
	case "delete_task":
		return handleDeleteTask(ctx, inputJSON)

	// Reminders
	case "list_reminders":
		return handleListReminders(ctx, inputJSON)
	case "create_reminder":
		return handleCreateReminder(ctx, inputJSON)
	case "delete_reminder":
		return handleDeleteReminder(ctx, inputJSON)

	// Log entries
	case "create_log_entry":
		return handleCreateLogEntry(ctx, inputJSON)
	case "list_log_entries":
		return handleListLogEntries(ctx, inputJSON)
	case "list_log_entries_by_tag":
		return handleListLogEntriesByTag(ctx, inputJSON)
	case "mood_trend":
		return handleMoodTrend(ctx, inputJSON)

	default:
		return jsonErr(fmt.Sprintf("unknown tool: %s", name)), nil
	}
}

// ============================================================
// Helpers
// ============================================================

// jsonOK marshals a result value to a JSON string for tool_result.
func jsonOK(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return jsonErr("failed to marshal result"), nil
	}
	return string(b), nil
}

// jsonErr returns a simple JSON error object. Never returns a Go error —
// tool errors are surfaced to the model as content, not panics.
func jsonErr(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}

// parseISO parses an ISO8601 datetime string. Returns zero time on empty input.
func parseISO(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse datetime %q", s)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// ============================================================
// Agenda
// ============================================================

func handleGetAgenda(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Period   string `json:"period"`
		FromDate string `json:"from_date"`
		ToDate   string `json:"to_date"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}

	var items []models.AgendaItem
	var err error

	switch in.Period {
	case "today":
		items, err = dal.TodayAgenda(ctx)
	case "week":
		items, err = dal.WeekAgenda(ctx)
	case "custom":
		from, e1 := parseISO(in.FromDate)
		to, e2 := parseISO(in.ToDate)
		if e1 != nil || e2 != nil {
			return jsonErr("invalid from_date or to_date"), nil
		}
		items, err = dal.GetAgenda(ctx, from, to)
	default:
		return jsonErr(`period must be "today", "week", or "custom"`), nil
	}

	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if items == nil {
		items = []models.AgendaItem{}
	}
	return jsonOK(items)
}

// ============================================================
// Projects
// ============================================================

func handleListProjects(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Status string `json:"status"`
	}
	json.Unmarshal(raw, &in)
	projects, err := dal.ListProjects(ctx, models.ProjectStatus(in.Status))
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if projects == nil {
		projects = []*models.Project{}
	}
	return jsonOK(projects)
}

func handleCreateProject(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Color       string `json:"color"`
		DueDate     string `json:"due_date"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	dueDate, _ := parseISO(in.DueDate)
	p := &models.Project{
		Name:        in.Name,
		Description: strPtr(in.Description),
		Status:      models.ProjectStatus(in.Status),
		Color:       strPtr(in.Color),
		DueDate:     timePtr(dueDate),
	}
	result, err := dal.CreateProject(ctx, p)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleUpdateProject(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Color       string `json:"color"`
		DueDate     string `json:"due_date"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	// Load existing so we only overwrite provided fields
	existing, err := dal.GetProject(ctx, in.ID)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if in.Name != "" {
		existing.Name = in.Name
	}
	if in.Description != "" {
		existing.Description = strPtr(in.Description)
	}
	if in.Status != "" {
		existing.Status = models.ProjectStatus(in.Status)
	}
	if in.Color != "" {
		existing.Color = strPtr(in.Color)
	}
	if in.DueDate != "" {
		d, _ := parseISO(in.DueDate)
		existing.DueDate = timePtr(d)
	}
	result, err := dal.UpdateProject(ctx, existing)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleDeleteProject(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(raw, &in)
	if err := dal.DeleteProject(ctx, in.ID); err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(map[string]any{"deleted": true, "id": in.ID})
}

// ============================================================
// Events
// ============================================================

func handleListEvents(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		FromDate string `json:"from_date"`
		ToDate   string `json:"to_date"`
	}
	json.Unmarshal(raw, &in)
	from, err := parseISO(in.FromDate)
	if err != nil {
		return jsonErr("invalid from_date"), nil
	}
	to, err := parseISO(in.ToDate)
	if err != nil {
		return jsonErr("invalid to_date"), nil
	}
	events, err := dal.ListEventsInRange(ctx, from, to)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if events == nil {
		events = []*models.Event{}
	}
	return jsonOK(events)
}

func handleCreateEvent(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		Location    string `json:"location"`
		AllDay      bool   `json:"all_day"`
		ProjectID   int64  `json:"project_id"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	start, err := parseISO(in.StartAt)
	if err != nil || start.IsZero() {
		return jsonErr("invalid start_at"), nil
	}
	end, err := parseISO(in.EndAt)
	if err != nil || end.IsZero() {
		return jsonErr("invalid end_at"), nil
	}
	e := &models.Event{
		Title:       in.Title,
		Description: strPtr(in.Description),
		StartAt:     start,
		EndAt:       end,
		Location:    strPtr(in.Location),
		AllDay:      in.AllDay,
		ProjectID:   int64Ptr(in.ProjectID),
	}
	result, err := dal.CreateEvent(ctx, e)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleUpdateEvent(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		Location    string `json:"location"`
		AllDay      *bool  `json:"all_day"`
		ProjectID   int64  `json:"project_id"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	existing, err := dal.GetEvent(ctx, in.ID)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if in.Title != "" {
		existing.Title = in.Title
	}
	if in.Description != "" {
		existing.Description = strPtr(in.Description)
	}
	if in.StartAt != "" {
		if t, err := parseISO(in.StartAt); err == nil {
			existing.StartAt = t
		}
	}
	if in.EndAt != "" {
		if t, err := parseISO(in.EndAt); err == nil {
			existing.EndAt = t
		}
	}
	if in.Location != "" {
		existing.Location = strPtr(in.Location)
	}
	if in.AllDay != nil {
		existing.AllDay = *in.AllDay
	}
	if in.ProjectID != 0 {
		existing.ProjectID = int64Ptr(in.ProjectID)
	}
	result, err := dal.UpdateEvent(ctx, existing)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleDeleteEvent(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(raw, &in)
	if err := dal.DeleteEvent(ctx, in.ID); err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(map[string]any{"deleted": true, "id": in.ID})
}

// ============================================================
// Tasks
// ============================================================

func handleListTasks(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Status    string `json:"status"`
		ProjectID int64  `json:"project_id"`
	}
	json.Unmarshal(raw, &in)
	tasks, err := dal.ListTasks(ctx, models.TaskStatus(in.Status), int64Ptr(in.ProjectID))
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if tasks == nil {
		tasks = []*models.Task{}
	}
	return jsonOK(tasks)
}

func handleGetTask(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(raw, &in)
	task, err := dal.GetTaskWithSubtasks(ctx, in.ID)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(task)
}

func handleListOverdueTasks(ctx context.Context, _ json.RawMessage) (string, error) {
	tasks, err := dal.ListOverdueTasks(ctx)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if tasks == nil {
		tasks = []*models.Task{}
	}
	return jsonOK(tasks)
}

func handleCreateTask(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    int    `json:"priority"`
		DueAt       string `json:"due_at"`
		ProjectID   int64  `json:"project_id"`
		ParentID    int64  `json:"parent_id"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	dueAt, _ := parseISO(in.DueAt)
	t := &models.Task{
		Title:       in.Title,
		Description: strPtr(in.Description),
		Status:      models.TaskStatus(in.Status),
		Priority:    models.TaskPriority(in.Priority),
		DueAt:       timePtr(dueAt),
		ProjectID:   int64Ptr(in.ProjectID),
		ParentID:    int64Ptr(in.ParentID),
	}
	result, err := dal.CreateTask(ctx, t)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleUpdateTask(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    int    `json:"priority"`
		DueAt       string `json:"due_at"`
		ProjectID   int64  `json:"project_id"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	existing, err := dal.GetTask(ctx, in.ID)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if in.Title != "" {
		existing.Title = in.Title
	}
	if in.Description != "" {
		existing.Description = strPtr(in.Description)
	}
	if in.Status != "" {
		existing.Status = models.TaskStatus(in.Status)
	}
	if in.Priority != 0 {
		existing.Priority = models.TaskPriority(in.Priority)
	}
	if in.DueAt != "" {
		d, _ := parseISO(in.DueAt)
		existing.DueAt = timePtr(d)
	}
	if in.ProjectID != 0 {
		existing.ProjectID = int64Ptr(in.ProjectID)
	}
	result, err := dal.UpdateTask(ctx, existing)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleUpdateTaskStatus(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	json.Unmarshal(raw, &in)
	result, err := dal.UpdateTaskStatus(ctx, in.ID, models.TaskStatus(in.Status))
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleDeleteTask(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(raw, &in)
	if err := dal.DeleteTask(ctx, in.ID); err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(map[string]any{"deleted": true, "id": in.ID})
}

// ============================================================
// Reminders
// ============================================================

func handleListReminders(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		IncludeInactive bool `json:"include_inactive"`
	}
	json.Unmarshal(raw, &in)

	var reminders []*models.Reminder
	var err error
	if in.IncludeInactive {
		// Fetch all reminders — active and inactive
		reminders, err = dal.ListActiveReminders(ctx) // TODO: add ListAllReminders to DAL if needed
	} else {
		reminders, err = dal.ListActiveReminders(ctx)
	}
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if reminders == nil {
		reminders = []*models.Reminder{}
	}
	return jsonOK(reminders)
}

func handleCreateReminder(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Message        string `json:"message"`
		RemindAt       string `json:"remind_at"`
		EventID        int64  `json:"event_id"`
		TaskID         int64  `json:"task_id"`
		RecurrenceType string `json:"recurrence_type"`
		RecurrenceRule string `json:"recurrence_rule"`
		RecurrenceEnd  string `json:"recurrence_end"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	remindAt, err := parseISO(in.RemindAt)
	if err != nil || remindAt.IsZero() {
		return jsonErr("invalid remind_at"), nil
	}
	recEnd, _ := parseISO(in.RecurrenceEnd)

	r := &models.Reminder{
		Message:        in.Message,
		RemindAt:       remindAt,
		EventID:        int64Ptr(in.EventID),
		TaskID:         int64Ptr(in.TaskID),
		RecurrenceRule: strPtr(in.RecurrenceRule),
		RecurrenceEnd:  timePtr(recEnd),
	}
	if in.RecurrenceType != "" {
		rt := models.RecurrenceType(in.RecurrenceType)
		r.RecurrenceType = &rt
	}
	result, err := dal.CreateReminder(ctx, r)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleDeleteReminder(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		ID int64 `json:"id"`
	}
	json.Unmarshal(raw, &in)
	if err := dal.DeleteReminder(ctx, in.ID); err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(map[string]any{"deleted": true, "id": in.ID})
}

// ============================================================
// Log entries
// ============================================================

func handleCreateLogEntry(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		EntryType        string   `json:"entry_type"`
		Title            string   `json:"title"`
		Body             string   `json:"body"`
		MoodScore        *int     `json:"mood_score"`
		Situation        string   `json:"situation"`
		AutomaticThought string   `json:"automatic_thought"`
		Reframe          string   `json:"reframe"`
		ProjectID        int64    `json:"project_id"`
		TaskID           int64    `json:"task_id"`
		Tags             []string `json:"tags"`
		EntryDate        string   `json:"entry_date"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return jsonErr("invalid input: " + err.Error()), nil
	}
	entryDate, _ := parseISO(in.EntryDate)
	e := &models.LogEntry{
		EntryType:        models.EntryType(in.EntryType),
		Title:            strPtr(in.Title),
		Body:             in.Body,
		MoodScore:        in.MoodScore,
		Situation:        strPtr(in.Situation),
		AutomaticThought: strPtr(in.AutomaticThought),
		Reframe:          strPtr(in.Reframe),
		ProjectID:        int64Ptr(in.ProjectID),
		TaskID:           int64Ptr(in.TaskID),
		Tags:             in.Tags,
		EntryDate:        entryDate,
	}
	result, err := dal.CreateLogEntry(ctx, e)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(result)
}

func handleListLogEntries(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		EntryType string `json:"entry_type"`
		FromDate  string `json:"from_date"`
		ToDate    string `json:"to_date"`
	}
	json.Unmarshal(raw, &in)
	from, _ := parseISO(in.FromDate)
	to, _ := parseISO(in.ToDate)
	entries, err := dal.ListLogEntries(ctx, models.EntryType(in.EntryType), from, to)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if entries == nil {
		entries = []*models.LogEntry{}
	}
	return jsonOK(entries)
}

func handleListLogEntriesByTag(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		Tag string `json:"tag"`
	}
	json.Unmarshal(raw, &in)
	entries, err := dal.ListLogEntriesByTag(ctx, in.Tag)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	if entries == nil {
		entries = []*models.LogEntry{}
	}
	return jsonOK(entries)
}

func handleMoodTrend(ctx context.Context, raw json.RawMessage) (string, error) {
	var in struct {
		FromDate string `json:"from_date"`
		ToDate   string `json:"to_date"`
	}
	json.Unmarshal(raw, &in)
	from, err := parseISO(in.FromDate)
	if err != nil {
		return jsonErr("invalid from_date"), nil
	}
	to, err := parseISO(in.ToDate)
	if err != nil {
		return jsonErr("invalid to_date"), nil
	}
	points, err := dal.MoodTrend(ctx, from, to)
	if err != nil {
		return jsonErr(err.Error()), nil
	}
	return jsonOK(points)
}
