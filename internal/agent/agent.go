package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// agent.go implements the agentic conversation loop:
//
//  1. Receive a user message
//  2. POST to Anthropic /v1/messages with tools + history
//  3. If the response contains tool_use blocks, dispatch each to DispatchTool
//  4. Append the assistant message and all tool_results to the conversation
//  5. POST again with the updated history (the model may call more tools)
//  6. Repeat until the model produces a stop_reason of "end_turn"
//  7. Return the final text response
//
// The loop is capped at maxToolRounds to prevent runaway billing.

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
	anthropicModel  = "claude-haiku-4-5-20251001"
	maxTokens       = 1024
	maxToolRounds   = 10
)

// buildSystemPrompt returns the system prompt with the current UTC time injected.
// Called fresh on every Reply() so the model always has an accurate "now".
func buildSystemPrompt() string {
	now := time.Now().UTC().Format("2006-01-02T15:04:05")
	return "You are a personal assistant managing a calendar, task list, and journal.\n" +
		"You have access to tools for creating and querying events, tasks, reminders, projects, and log entries.\n\n" +
		"Guidelines:\n" +
		"- Always use get_agenda before answering questions about what is on or what is coming up.\n" +
		"- When the user asks to add, create, schedule, or remind, call the appropriate write tool immediately without asking for confirmation unless a required field is genuinely missing.\n" +
		"- For datetime inputs always use ISO8601 format: 2026-05-03T14:00:00.\n" +
		"- When the user says today, tomorrow, or this week, resolve to concrete datetimes yourself based on the current date before calling tools.\n" +
		"- Keep responses concise. Use plain text with no markdown headers or bullet syntax, because this is WhatsApp.\n" +
		"- After a write operation, confirm what was created or updated in one sentence.\n" +
		"- Today's date and time (UTC): " + now + " UTC"
}

// ============================================================
// Anthropic API types
// ============================================================

// Message represents a single turn in the conversation history.
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Content is a polymorphic content block.
// Type is one of: "text", "tool_use", "tool_result".
type Content struct {
	Type string `json:"type"`

	// text block
	Text string `json:"text,omitempty"`

	// tool_use block (model -> us)
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// tool_result block (us -> model)
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
}

// apiRequest is the body sent to /v1/messages.
type apiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools"`
}

// apiResponse is the body returned by /v1/messages.
type apiResponse struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Role       string    `json:"role"`
	Content    []Content `json:"content"`
	StopReason string    `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ============================================================
// Agent
// ============================================================

// Agent holds the HTTP client and API key.
type Agent struct {
	client *http.Client
	apiKey string
}

// New creates a new Agent. API key is read from ANTHROPIC_API_KEY env var.
func New() (*Agent, error) {
	key := os.Getenv("XAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("XAI_API_KEY environment variable not set")
	}
	return &Agent{
		client: &http.Client{Timeout: 60 * time.Second},
		apiKey: key,
	}, nil
}

// ============================================================
// HTTP Handler
// ============================================================

type chatRequest struct {
	Message `json:"message"`
}

// take session/conversation id in request, load conversation history
// respond with agent reply
func (a *Agent) ReplyHandler(ctx context.Context, conv *Conversation) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// read request body
        log.Println("agent.ReplyHandler: ", r.URL.Host, r.URL.RequestURI())
		defer r.Body.Close()

		var cr chatRequest
		err := json.NewDecoder(r.Body).Decode(&cr)
		if err != nil {
			w.Write([]byte("error decoding request body"))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("agent.ReplyHandler: chat message :", cr)

		// load converation history
		
		// ask agent for reply
		// reply, err := a.Reply(context.Background(), conv, cr.Message)
		// if err != nil {
		// 	w.Write([]byte("error decoding request body"))
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }

		// save messages to conversation history

		// respond with agent reply
		// w.Write([]byte(reply))
		w.Write([]byte(`{ "message": "hi from the calendar agent" }`))
	}
}

// ============================================================
// Conversation
// ============================================================

// Conversation holds the message history for a single session.
// For WhatsApp integration, persist this between turns keyed by phone number.
type Conversation struct {
	History []Message
}

// NewConversation creates an empty conversation.
func NewConversation() *Conversation {
	return &Conversation{}
}

// Reply takes a user message, runs the agent loop, and returns the
// assistant's final text response.
func (a *Agent) Reply(ctx context.Context, conv *Conversation, userMessage string) (string, error) {
	// Append the new user turn
	conv.History = append(conv.History, Message{
		Role:    "user",
		Content: []Content{{Type: "text", Text: userMessage}},
	})

	// Agentic loop
	for round := 0; round < maxToolRounds; round++ {
		resp, err := a.callAPI(ctx, conv.History)
		if err != nil {
			return "", fmt.Errorf("API call: %w", err)
		}

		// Log token usage for cost monitoring on the Pi
		log.Printf("agent: round=%d stop=%s tokens(in=%d out=%d)",
			round, resp.StopReason, resp.Usage.InputTokens, resp.Usage.OutputTokens)

		// Append the full assistant message to history (required by the API)
		conv.History = append(conv.History, Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// If no tool calls, we are done
		if resp.StopReason == "end_turn" || !hasToolUse(resp.Content) {
			return extractText(resp.Content), nil
		}

		// Process all tool_use blocks in this response
		toolResults := make([]Content, 0, len(resp.Content))
		for _, block := range resp.Content {
			if block.Type != "tool_use" {
				continue
			}
			log.Printf("agent: tool_use name=%s id=%s", block.Name, block.ID)

			result, err := DispatchTool(ctx, block.Name, block.Input)
			if err != nil {
				result = jsonErr(err.Error())
			}

			log.Printf("agent: tool_result id=%s result_len=%d", block.ID, len(result))

			toolResults = append(toolResults, Content{
				Type:      "tool_result",
				ToolUseID: block.ID,
				Content:   result,
			})
		}

		// Append all tool results as a single user turn (Anthropic API requirement)
		conv.History = append(conv.History, Message{
			Role:    "user",
			Content: toolResults,
		})
	}

	return "", fmt.Errorf("agent: exceeded max tool rounds (%d)", maxToolRounds)
}

// ============================================================
// API call
// ============================================================

func (a *Agent) callAPI(ctx context.Context, history []Message) (*apiResponse, error) {
	reqBody := apiRequest{
		Model:     anthropicModel,
		MaxTokens: maxTokens,
		System:    buildSystemPrompt(),
		Messages:  history,
		Tools:     AllTools,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	httpResp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer httpResp.Body.Close()

	respBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", httpResp.StatusCode, string(respBytes))
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("API error [%s]: %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	return &apiResp, nil
}

// ============================================================
// Helpers
// ============================================================

func hasToolUse(content []Content) bool {
	for _, c := range content {
		if c.Type == "tool_use" {
			return true
		}
	}
	return false
}

func extractText(content []Content) string {
	var result string
	for _, c := range content {
		if c.Type == "text" {
			result += c.Text
		}
	}
	return result
}
