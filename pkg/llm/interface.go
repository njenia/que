package llm

import (
	"fmt"

	"github.com/jenian/que/internal/config"
)

// Client is the common interface for LLM providers
type Client interface {
	// QueryWithPayload queries the LLM with a structured payload for initial analysis
	// Returns a JSON response with root_cause, evidence, and fix fields
	QueryWithPayload(cfg *config.Config, payload config.QueryPayload) (string, error)

	// QueryWithHistory queries the LLM with conversation history for interactive mode
	// conversationHistory is a flat array: [user1, assistant1, user2, assistant2, ...]
	// Returns a plain text response (not JSON)
	QueryWithHistory(cfg *config.Config, conversationHistory []string, userQuestion string) (string, error)
}

// NewClient creates a new LLM client based on the provider specified in config
func NewClient(cfg *config.Config) (Client, error) {
	switch cfg.Provider {
	case "openai":
		return NewOpenAIClientFromConfig(cfg)
	case "claude":
		return NewAnthropicClientFromConfig(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

