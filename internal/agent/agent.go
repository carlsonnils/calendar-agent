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

	"fake.com/nilspcarlson/internal/dal"
)

// agent.go implements the agentic conversation loop:
//
//  1. Receive a user message
//  2. POST to xAI chat endpoint
//  3. If the response contains tool blocks, dispatch each to DispatchTool
//  4. Append the assistant message and all tool_results to the conversation
//  5. POST again with the updated history (the model may call more tools)
//  6. Repeat until the model produces a stop_reason of "end_turn"
//  7. Return the final text response
//
// The loop is capped at maxToolRounds to prevent runaway billing.

const (
	xaiAPIURL = "https://api.x.ai/v1/chat/completions"
	xaiModel  = "grok-4-1-fast-reasoning"
	maxCompletionTokens       = 1024
	maxToolRounds   = 5
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
		"- Keep responses concise. Use html when it makes sense to format responses. Your response will be embeded into a web page.\n" +
		"- After a write operation, confirm what was created or updated in one sentence.\n" +
		"- Today's date and time (UTC): " + now + " UTC"
}

// ============================================================
// Anthropic API types
// ============================================================

// Message represents a single turn in the conversation history.
type Message struct {
	Role    string    `json:"role"`
	Content string `json:"content"`

	// for assistant messages
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// for tool messages
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// apiRequest is the body sent to /v1/chat/completions.
type apiRequest struct {
	Model     string    `json:"model"`
	MaxCompletionTokens int       `json:"max_completion_tokens"`
	Messages  []Message `json:"messages"`
	Tools     []Tool    `json:"tools"`
}

// apiResponse tool call Fucntion
type ToolCallFunction struct {
	Arguments string `json:"arguments"`
	Name string `json:"name"`
}

// apiResponse Tool Call
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
	ID string `json:"id"`
	Index int `json:"index,omitempty"`
	Type string `json:"type,omitempty"`
}

// apiResponce choice message
type ChoiceMessage struct {
	Content string `json:"content,omitempty"`
	Role string `json:"role"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// apiResponse Choices type
type Choice struct {
	FinishReason string `json:"finish_reason,omitempty"`
	Index int `json:"index"`
	Message ChoiceMessage `json:"message"`
}

// apiResponse is the body returned by /v1/chat/completions.
type apiResponse struct {
	Choices []Choice `json:"choices`
	Created int `json:"created"`
	ID string `json:"id"`
	Model string `json:"model"`
	Object string `json:"object"`
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
	Message string `json:"message"`
}

type chatResponse struct {
	Message string `json:"message"`
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
			log.Println("agent.ReplyHandler error: ", err)
			json.NewEncoder(w).Encode(&chatRequest{Message: "error decoding request body"})
			return
		}
		log.Println("agent.ReplyHandler: chat message :", cr)

		// ask agent for reply
		reply, err := a.Reply(context.Background(), conv, cr.Message)
		if err != nil {
			log.Println("agent.ReplyHandler error: ", err)
			json.NewEncoder(w).Encode(&chatRequest{Message: "error getting reply"})
			return
		}

		// respond with agent reply
		json.NewEncoder(w).Encode(&chatRequest{Message: reply})
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

func LoadConversation(sessionId string) (*Conversation, error) {
	c, err := dal.LoadConversation(context.Background(), sessionId)
	if err != nil {
		return NewConversation(), err
	}

	conv := &Conversation{}
	err = json.Unmarshal(c.History, &conv.History)
	if err != nil {
		return NewConversation(), err
	}

	return conv, nil
}

func saveConversation(conv *Conversation) error {
	// get history bytes
	histBytes, err := json.Marshal(conv.History)
	if err != nil {
		return err
	} 

	err = dal.SaveConversation(
		context.Background(), "nilspcarlson_calendar_0", 
		"main calendar", histBytes, len(conv.History))
	if err != nil {
		return err
	}

	return nil
}

func processToolCalls(choice Choice, conv *Conversation, ctx context.Context) {
	for _, tc := range choice.Message.ToolCalls {
		log.Printf("agent: tool_use name=%s id=%s",  tc.Function.Name, tc.ID)
		result, err := DispatchTool(ctx, tc.Function.Name, []byte(tc.Function.Arguments))
		if err != nil {
			result = jsonErr(err.Error())
		}

		// Append all tool results as a single user turn
		conv.History = append(conv.History, Message{
			Role:    "tool",
			Content: result,
			ToolCallID: tc.ID,
		})
	}
}

// Reply takes a user message, runs the agent loop, and returns the
// assistant's final text response.
func (a *Agent) Reply(ctx context.Context, conv *Conversation, userMessage string) (string, error) {
	// save conversation on function exit
	defer func() {
		err := saveConversation(conv)
		if err != nil {
			log.Println("agent.Reply error saving conversation:", err)
		}
	}()

	// Append the new user turn
	conv.History = append(conv.History, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Agentic loop
	for round := 0; round < maxToolRounds; round++ {
		// use only most recent 10 messages for conversation context history
		hist := conv.History
		if len(hist) > 10 {
			hist = hist[len(hist)-10:len(hist)]
		}

		// call chat xAI endpoint with the conversation history
		resp, err := a.callAPI(ctx, hist)
		if err != nil {
			return "", err
		}

		// only use first choice, not testing llm outputs
		choice := resp.Choices[0]

		// Log token usage for cost monitoring
		log.Printf("agent: round=%d stop=%s",
			round, choice.FinishReason)
		log.Println("agent.Reply llm response choices: ", resp.Choices)

		// Append the full assistant message to history
		conv.History = append(conv.History, Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
			ToolCalls: choice.Message.ToolCalls,
		})

		// If no tool calls return from function
		if choice.FinishReason == "stop" || choice.FinishReason != "tool_calls" {
			return choice.Message.Content, nil
		}

		// Process all tool_use blocks
		processToolCalls(choice, conv, ctx)
	}

	return "", fmt.Errorf("agent: exceeded max tool rounds (%d)", maxToolRounds)
}

// ============================================================
// API call
// ============================================================

func (a *Agent) callAPI(ctx context.Context, history []Message) (*apiResponse, error) {
	reqBody := apiRequest{
		Model:     xaiModel,
		MaxCompletionTokens: maxCompletionTokens,
		Messages:  append(history, Message{
			Role: "system",
			Content: buildSystemPrompt(),
		}),
		Tools:     AllTools,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, xaiAPIURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

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

	return &apiResp, nil
}

// ============================================================
// Helpers
// ============================================================

// func extractText(choices []Choice) string {
// 	var result string
// 	for _, c := range choices {
// 		result += c.Message.Content
// 	}
// 	return result
// }
