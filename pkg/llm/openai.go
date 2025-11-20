package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/jenian/que/internal/config"
	"github.com/sashabaranov/go-openai"
)

// OpenAIClient handles interactions with OpenAI API
type OpenAIClient struct {
	client *openai.Client
	model  string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey string, modelOverride string) (*OpenAIClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(apiKey)
	
	model := "gpt-4o"
	if modelOverride != "" {
		model = modelOverride
	}

	return &OpenAIClient{
		client: client,
		model:  model,
	}, nil
}

// Query sends a query to OpenAI and returns the response
func (c *OpenAIClient) Query(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// QueryWithPayload implements the Client interface
func (c *OpenAIClient) QueryWithPayload(cfg *config.Config, payload config.QueryPayload) (string, error) {
	// Format system prompt
	systemPrompt := "You are a CLI debugging assistant. You must respond with valid JSON only. The response must contain exactly four fields: status, root_cause, evidence, and fix. Do not include markdown, code blocks, or any text outside the JSON."
	
	// Format user prompt with context
	userPrompt := formatPrompt(payload)

	// Show full prompt in verbose mode
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
func (c *OpenAIClient) QueryWithHistory(cfg *config.Config, conversationHistory []string, userQuestion string) (string, error) {
	// Build conversation messages
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a CLI debugging assistant. Provide concise, helpful answers. Keep responses brief since we're in a terminal.",
		},
	}

	// Add conversation history
	for i, msg := range conversationHistory {
		if i%2 == 0 {
			// User message
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg,
			})
		} else {
			// Assistant message
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg,
			})
		}
	}

	// Add current question
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userQuestion,
	})

	ctx := context.Background()
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// NewOpenAIClientFromConfig creates a new OpenAI client from config
func NewOpenAIClientFromConfig(cfg *config.Config) (Client, error) {
	return NewOpenAIClient(cfg.ChatGPTKey, cfg.Model)
}

