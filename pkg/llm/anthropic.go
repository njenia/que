package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jenian/que/internal/config"
)

const (
	anthropicAPIURL = "https://api.anthropic.com/v1/messages"
)

// AnthropicClient handles interactions with Anthropic API
type AnthropicClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(apiKey string, modelOverride string) (*AnthropicClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	model := "claude-3-5-sonnet-20241022"
	if modelOverride != "" {
		model = modelOverride
	}

	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// anthropicRequest represents the request body for Anthropic API
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the response from Anthropic API
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *anthropicError `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Query sends a query to Anthropic and returns the response
func (c *AnthropicClient) Query(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		Messages: []message{
			{
				Role:    "user",
				Content: fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt),
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr anthropicResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != nil {
			return "", fmt.Errorf("anthropic API error: %s - %s", apiErr.Error.Type, apiErr.Error.Message)
		}
		return "", fmt.Errorf("anthropic API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return apiResp.Content[0].Text, nil
}

// QueryWithPayload implements the Client interface
func (c *AnthropicClient) QueryWithPayload(cfg *config.Config, payload config.QueryPayload) (string, error) {
	// Format system prompt
	systemPrompt := "You are a CLI debugging assistant. You must respond with valid JSON only. The response must contain exactly four fields: status, root_cause, evidence, and fix. Do not include markdown, code blocks, or any text outside the JSON."

	// Format user prompt with context
	userPrompt := formatPrompt(payload)

	// Show full prompt in verbose mode (Anthropic combines system + user in user message)
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "\n=== LLM Prompt ===\n")
		fmt.Fprintf(os.Stderr, "System Prompt:\n%s\n\n", systemPrompt)
		fmt.Fprintf(os.Stderr, "User Prompt:\n%s\n\n", userPrompt)
		fmt.Fprintf(os.Stderr, "=== End Prompt ===\n\n")
	}

	ctx := context.Background()
	response, err := c.Query(ctx, systemPrompt, userPrompt)

	// Show raw response in verbose mode
	if cfg.Verbose && err == nil {
		fmt.Fprintf(os.Stderr, "=== LLM Raw Response ===\n%s\n=== End Response ===\n\n", response)
	}

	return response, err
}

// QueryWithHistory implements the Client interface
func (c *AnthropicClient) QueryWithHistory(cfg *config.Config, conversationHistory []string, userQuestion string) (string, error) {
	// Build conversation messages
	// Anthropic uses a different format - system prompt is included in first user message
	systemPrompt := "You are a CLI debugging assistant. Provide concise, helpful answers. Keep responses brief since we're in a terminal."

	var messages []message

	// Add conversation history
	firstMessage := true
	for i, msg := range conversationHistory {
		if i%2 == 0 {
			// User message
			content := msg
			if firstMessage {
				content = systemPrompt + "\n\n" + msg
				firstMessage = false
			}
			messages = append(messages, message{
				Role:    "user",
				Content: content,
			})
		} else {
			// Assistant message
			messages = append(messages, message{
				Role:    "assistant",
				Content: msg,
			})
		}
	}

	// Add current question (with system prompt if no history)
	if firstMessage {
		userQuestion = systemPrompt + "\n\n" + userQuestion
	}
	messages = append(messages, message{
		Role:    "user",
		Content: userQuestion,
	})

	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 2048, // Shorter for follow-ups
		Messages:  messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr anthropicResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error != nil {
			return "", fmt.Errorf("anthropic API error: %s - %s", apiErr.Error.Type, apiErr.Error.Message)
		}
		return "", fmt.Errorf("anthropic API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return apiResp.Content[0].Text, nil
}

// NewAnthropicClientFromConfig creates a new Anthropic client from config
func NewAnthropicClientFromConfig(cfg *config.Config) (Client, error) {
	return NewAnthropicClient(cfg.ClaudeKey, cfg.Model)
}
