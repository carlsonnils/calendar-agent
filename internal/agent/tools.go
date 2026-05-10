package agent

// tools.go defines every tool the agent can call.
//
// Design principles:
//   - Each tool maps 1:1 to a DAL function — no business logic in definitions.
//   - Required fields are the minimum needed for a valid DB write.
//   - Descriptions are written for Haiku: specific, action-oriented, include
//     example values so the model picks the right types without guessing.
//   - Datetimes are always ISO8601 strings ("2026-05-03T14:00:00") — we parse
//     in the handler, not in the tool definition.

// Tool represents a single Anthropic API tool definition.
type XAIToolType string

var (
	XAIFunctionToolType XAIToolType = "function"
)

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters ToolParameters `json:"parameters"`
}

type Tool struct {
	Function ToolFunction `json:"function"`
	Type XAIToolType `json:"type"`
}

// ToolParameters is the JSON Schema object for tool input validation.
type ToolParameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property is a single JSON Schema property definition.
type Property struct {
	Type        string   `json:"type,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Items       *Items   `json:"items,omitempty"` // for array types
}

// Items is used inside array-typed properties.
type Items struct {
	Type string `json:"type"`
}

// AllTools is the complete tool list sent to the Anthropic API on every request.
var AllTools = []Tool{

	// ============================================================
	// AGENDA — cross-cutting read
	// ============================================================
	{
		Function: ToolFunction{
			Name: "get_agenda",
			Description: `Return a unified, time-ordered list of events, tasks, and reminders
for a given time window. Use this for queries like "what's on today",
"what do I have this week", "show me everything for the next 3 days".
For "today" pass period="today". For "this week" pass period="week".
For a custom range pass period="custom" with from_date and to_date.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"period": {
						Type:        "string",
						Description: `Shorthand period. Use "today", "week", or "custom".`,
						Enum:        []string{"today", "week", "custom"},
					},
					"from_date": {
						Type:        "string",
						Description: `Start of custom range. ISO8601 datetime e.g. "2026-05-03T00:00:00". Required when period="custom".`,
					},
					"to_date": {
						Type:        "string",
						Description: `End of custom range. ISO8601 datetime e.g. "2026-05-10T23:59:59". Required when period="custom".`,
					},
				},
				Required: []string{"period"},
			},
		},
		Type: XAIFunctionToolType,
	},

	// ============================================================
	// PROJECTS
	// ============================================================
	{
		Function: ToolFunction{
			Name:        "list_projects",
			Description: `List all projects. Optionally filter by status: "active", "paused", "completed", or "archived". Omit status to return all.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"status": {
						Type:        "string",
						Description: `Filter by project status. Omit to return all projects.`,
						Enum:        []string{"active", "paused", "completed", "archived"},
					},
				},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "create_project",
			Description: `Create a new project. Projects group related tasks and events together.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"name":        {Type: "string", Description: `Project name e.g. "Home Renovation".`},
					"description": {Type: "string", Description: `Optional longer description.`},
					"status": {
						Type:        "string",
						Description: `Project status. Defaults to "active".`,
						Enum:        []string{"active", "paused", "completed", "archived"},
					},
					"color":    {Type: "string", Description: `Optional hex color for display e.g. "#4A90E2".`},
					"due_date": {Type: "string", Description: `Optional due date. ISO8601 date e.g. "2026-08-31".`},
				},
				Required: []string{"name"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "update_project",
			Description: `Update an existing project. Only provide the fields you want to change.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"id":          {Type: "integer", Description: `Project ID to update.`},
					"name":        {Type: "string", Description: `New project name.`},
					"description": {Type: "string", Description: `New description.`},
					"status": {
						Type:        "string",
						Description: `New status.`,
						Enum:        []string{"active", "paused", "completed", "archived"},
					},
					"color":    {Type: "string", Description: `New hex color.`},
					"due_date": {Type: "string", Description: `New due date. ISO8601 date e.g. "2026-08-31".`},
				},
				Required: []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "delete_project",
			Description: `Delete a project by ID. Tasks and events linked to this project will have their project association removed but will NOT be deleted.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{"id": {Type: "integer", Description: `Project ID to delete.`}},
				Required:   []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},

	// ============================================================
	// EVENTS
	// ============================================================
	{
		Function: ToolFunction{
			Name:        "list_events",
			Description: `List calendar events within a date range. Returns events whose start time falls in [from_date, to_date].`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"from_date": {Type: "string", Description: `Range start. ISO8601 datetime e.g. "2026-05-03T00:00:00".`},
					"to_date":   {Type: "string", Description: `Range end. ISO8601 datetime e.g. "2026-05-10T23:59:59".`},
				},
				Required: []string{"from_date", "to_date"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "create_event",
			Description: `Create a new calendar event. Use this for appointments, meetings, deadlines with a specific time block.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"title":       {Type: "string", Description: `Event title e.g. "Dentist Appointment".`},
					"description": {Type: "string", Description: `Optional notes or details.`},
					"start_at":    {Type: "string", Description: `Start datetime. ISO8601 e.g. "2026-05-10T09:00:00".`},
					"end_at":      {Type: "string", Description: `End datetime. ISO8601 e.g. "2026-05-10T10:00:00".`},
					"location":    {Type: "string", Description: `Optional location or meeting link.`},
					"all_day":     {Type: "boolean", Description: `Set true for all-day events like holidays. Defaults to false.`},
					"project_id":  {Type: "integer", Description: `Optional project ID to associate this event with.`},
				},
				Required: []string{"title", "start_at", "end_at"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "update_event",
			Description: `Update an existing calendar event. Only provide the fields you want to change.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"id":          {Type: "integer", Description: `Event ID to update.`},
					"title":       {Type: "string", Description: `New title.`},
					"description": {Type: "string", Description: `New description.`},
					"start_at":    {Type: "string", Description: `New start datetime. ISO8601.`},
					"end_at":      {Type: "string", Description: `New end datetime. ISO8601.`},
					"location":    {Type: "string", Description: `New location.`},
					"all_day":     {Type: "boolean", Description: `Set all-day flag.`},
					"project_id":  {Type: "integer", Description: `New project association.`},
				},
				Required: []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "delete_event",
			Description: `Delete a calendar event by ID. Any reminders attached to this event will also be deleted.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{"id": {Type: "integer", Description: `Event ID to delete.`}},
				Required:   []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},

	// ============================================================
	// TASKS
	// ============================================================
	{
		Function: ToolFunction{
			Name: "list_tasks",
			Description: `List tasks. Optionally filter by status and/or project. Only returns top-level tasks (not subtasks).
	Status values: "todo", "in_progress", "blocked", "done", "cancelled".`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"status": {
						Type:        "string",
						Description: `Filter by task status. Omit to return all statuses.`,
						Enum:        []string{"todo", "in_progress", "blocked", "done", "cancelled"},
					},
					"project_id": {Type: "integer", Description: `Filter by project ID. Omit to return tasks from all projects.`},
				},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "get_task",
			Description: `Get a single task by ID, including its subtasks.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{"id": {Type: "integer", Description: `Task ID.`}},
				Required:   []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "list_overdue_tasks",
			Description: `List all tasks that are past their due date and not yet done or cancelled.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "create_task",
			Description: `Create a new task or subtask. For subtasks, provide parent_id. Priority: 1=high, 2=medium, 3=low.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"title":       {Type: "string", Description: `Task title e.g. "Write Q2 report".`},
					"description": {Type: "string", Description: `Optional details.`},
					"status": {
						Type:        "string",
						Description: `Initial status. Defaults to "todo".`,
						Enum:        []string{"todo", "in_progress", "blocked", "done", "cancelled"},
					},
					"priority":   {Type: "integer", Description: `1=high, 2=medium, 3=low. Defaults to 2.`},
					"due_at":     {Type: "string", Description: `Optional due datetime. ISO8601 e.g. "2026-05-15T17:00:00".`},
					"project_id": {Type: "integer", Description: `Optional project ID.`},
					"parent_id":  {Type: "integer", Description: `Optional parent task ID for creating a subtask.`},
				},
				Required: []string{"title"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "update_task",
			Description: `Update an existing task. Only provide the fields you want to change. To quickly mark a task done, use update_task_status instead.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"id":          {Type: "integer", Description: `Task ID to update.`},
					"title":       {Type: "string", Description: `New title.`},
					"description": {Type: "string", Description: `New description.`},
					"status": {
						Type: "string",
						Enum: []string{"todo", "in_progress", "blocked", "done", "cancelled"},
					},
					"priority":   {Type: "integer", Description: `1=high, 2=medium, 3=low.`},
					"due_at":     {Type: "string", Description: `New due datetime. ISO8601.`},
					"project_id": {Type: "integer", Description: `New project ID.`},
				},
				Required: []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "update_task_status",
			Description: `Quickly change only the status of a task. Use this for "mark X as done", "start working on Y", "block Z".`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"id": {Type: "integer", Description: `Task ID.`},
					"status": {
						Type:        "string",
						Description: `New status.`,
						Enum:        []string{"todo", "in_progress", "blocked", "done", "cancelled"},
					},
				},
				Required: []string{"id", "status"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{ 
			Name:        "delete_task",
			Description: `Delete a task by ID. Subtasks of this task will also be deleted.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{"id": {Type: "integer", Description: `Task ID to delete.`}},
				Required:   []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},

	// ============================================================
	// REMINDERS
	// ============================================================
	{
		Function: ToolFunction{
			Name: "list_reminders",
			Description: `List reminders. By default returns only active reminders.
	Set include_inactive=true to include fired/deactivated reminders.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"include_inactive": {Type: "boolean", Description: `Include inactive/fired reminders. Defaults to false.`},
				},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{ 
			Name: "create_reminder",
			Description: `Create a reminder. Can be one-shot or recurring.
	For recurring reminders set recurrence_type ("minutely","hourly","daily","weekly","monthly","yearly").
	For complex rules (e.g. every Monday and Wednesday) set recurrence_rule as a JSON string:
	  - Every 2 weeks: {"interval": 2}
	  - Specific days: {"interval": 1, "days": ["mon","wed","fri"]}
	  - Day of month:  {"interval": 1, "day_of_month": 15}
	Omit recurrence_type for a one-shot reminder.
	Optionally attach to an event (event_id) or task (task_id), but not both.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"message":   {Type: "string", Description: `The reminder message text.`},
					"remind_at": {Type: "string", Description: `When to fire. ISO8601 datetime e.g. "2026-05-10T08:00:00".`},
					"event_id":  {Type: "integer", Description: `Optional: attach to this event ID.`},
					"task_id":   {Type: "integer", Description: `Optional: attach to this task ID.`},
					"recurrence_type": {
						Type:        "string",
						Description: `Set for recurring reminders.`,
						Enum:        []string{"minutely", "hourly", "daily", "weekly", "monthly", "yearly", "custom"},
					},
					"recurrence_rule": {Type: "string", Description: `JSON string with recurrence details. e.g. '{"interval":1,"days":["mon","fri"]}'.`},
					"recurrence_end":  {Type: "string", Description: `Optional end date for recurrence. ISO8601 date e.g. "2026-12-31". Omit for indefinite.`},
				},
				Required: []string{"message", "remind_at"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "delete_reminder",
			Description: `Delete a reminder by ID.`,
			Parameters: ToolParameters{
				Type:       "object",
				Properties: map[string]Property{"id": {Type: "integer", Description: `Reminder ID to delete.`}},
				Required:   []string{"id"},
			},
		},
		Type: XAIFunctionToolType,
	},

	// ============================================================
	// LOG ENTRIES
	// ============================================================
	{
		Function: ToolFunction{
			Name: "create_log_entry",
			Description: `Create a log entry. Use entry_type to categorise:
	  "note"      — general thought or observation
	  "diary"     — narrative journal entry (end of day, reflection)
	  "work_log"  — what you worked on, progress, blockers
	  "thought"   — CBT-style thought record (use situation, automatic_thought, reframe fields)
	  "mood"      — quick mood snapshot (use mood_score 1-10)
	  "idea"      — capture an idea for later (not yet a task)
	  "win"       — log a win or accomplishment
	  "gratitude" — gratitude entry

	For "thought" type, populate situation, automatic_thought, and reframe.
	For mood tracking, set mood_score (1=very low, 10=excellent).
	Tags are a list of strings e.g. ["work","health"].`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"entry_type": {
						Type:        "string",
						Description: `Type of log entry.`,
						Enum:        []string{"note", "diary", "work_log", "thought", "mood", "idea", "win", "gratitude"},
					},
					"title":             {Type: "string", Description: `Optional headline. Leave blank and the content will speak for itself.`},
					"body":              {Type: "string", Description: `The main content of the entry.`},
					"mood_score":        {Type: "integer", Description: `Mood score 1-10. 1=very low, 5=neutral, 10=excellent.`},
					"situation":         {Type: "string", Description: `For thought records: what happened / the trigger.`},
					"automatic_thought": {Type: "string", Description: `For thought records: the immediate negative thought.`},
					"reframe":           {Type: "string", Description: `For thought records: the rational, balanced reframe.`},
					"project_id":        {Type: "integer", Description: `Optional: link this entry to a project.`},
					"task_id":           {Type: "integer", Description: `Optional: link this entry to a task.`},
					"tags": {
						Type:        "array",
						Description: `Optional list of string tags e.g. ["work","health","golf"].`,
						Items:       &Items{Type: "string"},
					},
					"entry_date": {Type: "string", Description: `When the entry is logically "about". Defaults to now. Use for backdating: "2026-05-01T22:00:00".`},
				},
				Required: []string{"entry_type", "body"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name: "list_log_entries",
			Description: `List log entries. Filter by entry_type and/or date range.
	Omit filters to return all recent entries.`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"entry_type": {
						Type:        "string",
						Description: `Filter by entry type.`,
						Enum:        []string{"note", "diary", "work_log", "thought", "mood", "idea", "win", "gratitude"},
					},
					"from_date": {Type: "string", Description: `Start of date range. ISO8601 datetime.`},
					"to_date":   {Type: "string", Description: `End of date range. ISO8601 datetime.`},
				},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "list_log_entries_by_tag",
			Description: `List log entries that have a specific tag e.g. "golf", "work", "health".`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"tag": {Type: "string", Description: `Tag to filter by e.g. "golf".`},
				},
				Required: []string{"tag"},
			},
		},
		Type: XAIFunctionToolType,
	},
	{
		Function: ToolFunction{
			Name:        "mood_trend",
			Description: `Return mood scores over a date range for tracking emotional trends. Use for "how has my mood been this month".`,
			Parameters: ToolParameters{
				Type: "object",
				Properties: map[string]Property{
					"from_date": {Type: "string", Description: `Start date. ISO8601 datetime.`},
					"to_date":   {Type: "string", Description: `End date. ISO8601 datetime.`},
				},
				Required: []string{"from_date", "to_date"},
			},
		},
		Type: XAIFunctionToolType,
	},
}
