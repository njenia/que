package config

import (
	"encoding/json"
	"strings"
	"time"
)

// Context represents system environment information
type Context struct {
	OS        string
	Arch      string
	Shell     string
	Timestamp time.Time
}

// QueryPayload contains the data to be sent to the LLM
type QueryPayload struct {
	RawLog        string
	SanitizedLog  string
	SystemContext Context
}

// Redactor interface allows swapping redaction strategies
type Redactor interface {
	Redact(input string) (string, int)                                           // returns redacted string and count of secrets found
	RedactWithDetails(input string, verbose bool) (string, int, []FindingDetail) // returns redacted string, count, and details for verbose output
}

// FindingDetail contains information about a detected secret for verbose output
type FindingDetail struct {
	RuleID      string
	Description string
	Match       string
	Secret      string
	Source      string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	Entropy     float32
	File        string
	Tags        []string
	Fingerprint string
}

// EvidenceString can unmarshal from both a string or an array of strings
type EvidenceString string

// UnmarshalJSON handles both string and array formats for evidence
func (e *EvidenceString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*e = EvidenceString(str)
		return nil
	}

	// If that fails, try as array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*e = EvidenceString(strings.Join(arr, "\n"))
		return nil
	}

	return json.Unmarshal(data, (*string)(e))
}

// LLMResponse represents the structured JSON response from the LLM
type LLMResponse struct {
	Status    string        `json:"status"` // "no_problem", "problem_detected", "insufficient_data"
	RootCause string        `json:"root_cause"`
	Evidence  EvidenceString `json:"evidence"`
	Fix       string        `json:"fix"`
}

// Config holds CLI flags and environment variables
type Config struct {
	Provider        string // "openai" or "claude"
	Model           string // Model override (optional)
	Verbose         bool
	NoContext       bool
	DryRun          bool
	Interactive     bool
	ChatGPTKey      string
	ClaudeKey       string
	DefaultProvider string
}

// NewConfig creates a new Config with defaults
func NewConfig() *Config {
	return &Config{
		Provider:        "openai",
		DefaultProvider: "openai",
	}
}
