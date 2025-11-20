package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jenian/que/internal/advisor"
	"github.com/jenian/que/internal/config"
	"github.com/jenian/que/internal/enricher"
	"github.com/jenian/que/internal/ingestor"
	"github.com/jenian/que/internal/sanitizer"
	"github.com/jenian/que/pkg/llm"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	// Example: go build -ldflags="-X main.Version=v1.0.0" ./cmd/que
	Version = "dev"
)

var (
	providerFlag    string
	modelFlag       string
	verboseFlag     bool
	noContextFlag   bool
	dryRunFlag      bool
	interactiveFlag bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "que",
		Short:   "The pipe-able DevOps assistant",
		Long:    fmt.Sprintf("Que is a CLI utility that analyzes logs and errors from stdin using LLMs to suggest fixes.\n\nVersion: %s", Version),
		Version: Version,
		RunE:    runQue,
	}

	rootCmd.Flags().StringVarP(&providerFlag, "provider", "p", "", "LLM provider to use (openai, claude)")
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Specific model override (e.g., gpt-4-turbo)")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Show what data is being sent (including redaction)")
	rootCmd.Flags().BoolVar(&noContextFlag, "no-context", false, "Skip environment context gathering")
	rootCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Perform redaction and context gathering but do not call API")
	rootCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "Enter interactive mode for follow-up questions")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runQue(cmd *cobra.Command, args []string) error {
	// Display header
	printHeader()

	// Load configuration
	cfg := config.NewConfig()

	// Load environment variables
	cfg.ChatGPTKey = os.Getenv("QUE_CHATGPT_API_KEY")
	cfg.ClaudeKey = os.Getenv("QUE_CLAUDE_API_KEY")
	if defaultProvider := os.Getenv("QUE_DEFAULT_PROVIDER"); defaultProvider != "" {
		cfg.DefaultProvider = defaultProvider
	}

	// Apply CLI flags
	if providerFlag != "" {
		cfg.Provider = providerFlag
	} else {
		cfg.Provider = cfg.DefaultProvider
	}
	cfg.Model = modelFlag
	cfg.Verbose = verboseFlag
	cfg.NoContext = noContextFlag
	cfg.DryRun = dryRunFlag
	cfg.Interactive = interactiveFlag

	// Validate provider
	if cfg.Provider != "openai" && cfg.Provider != "claude" {
		return fmt.Errorf("invalid provider: %s (must be 'openai' or 'claude')", cfg.Provider)
	}

	// Validate API key
	if !cfg.DryRun {
		if cfg.Provider == "openai" && cfg.ChatGPTKey == "" {
			return fmt.Errorf("QUE_CHATGPT_API_KEY environment variable is required for OpenAI provider")
		}
		if cfg.Provider == "claude" && cfg.ClaudeKey == "" {
			return fmt.Errorf("QUE_CLAUDE_API_KEY environment variable is required for Claude provider")
		}
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "Provider: %s\n", cfg.Provider)
		if cfg.Model != "" {
			fmt.Fprintf(os.Stderr, "Model: %s\n", cfg.Model)
		}
	}

	// Pipeline: Ingestor → Enricher → Sanitizer → Advisor
	rawLog, err := ingestor.Ingest()
	if err != nil {
		return fmt.Errorf("failed to ingest input: %w", err)
	}

	if len(rawLog) == 0 {
		return fmt.Errorf("no input provided on stdin")
	}

	var ctx config.Context
	if !cfg.NoContext {
		ctx = enricher.Enrich()
		if cfg.Verbose {
			fmt.Fprintf(os.Stderr, "System Context: OS=%s, Arch=%s, Shell=%s\n", ctx.OS, ctx.Arch, ctx.Shell)
		}
	}

	redactor := sanitizer.NewRedactor()
	var sanitizedLog string
	var redactionCount int

	if cfg.Verbose {
		// In verbose mode, we still redact but don't show the count message
		sanitizedLog, _ = redactor.Redact(rawLog)
	} else {
		sanitizedLog, redactionCount = redactor.Redact(rawLog)
		if redactionCount > 0 {
			fmt.Fprintf(os.Stderr, "Redacted %d potential secrets\n", redactionCount)
		}
	}

	payload := config.QueryPayload{
		RawLog:        rawLog,
		SanitizedLog:  sanitizedLog,
		SystemContext: ctx,
	}

	// Create LLM client (only if not in dry-run mode)
	var llmClient llm.Client
	if !cfg.DryRun {
		var err error
		llmClient, err = llm.NewClient(cfg)
		if err != nil {
			return fmt.Errorf("failed to create LLM client: %w", err)
		}
	}

	// Call advisor
	response, err := advisor.Advise(llmClient, cfg, payload)
	if err != nil {
		return fmt.Errorf("failed to get advice: %w", err)
	}

	// Output response
	fmt.Print(response)

	// Handle interactive mode (only if problems were detected)
	// Skip interactive mode if the response indicates no problems
	noProblemsDetected := strings.Contains(response, "no problems detected")
	if cfg.Interactive && !cfg.DryRun && !noProblemsDetected {
		return advisor.RunInteractive(llmClient, cfg, payload, response)
	}

	return nil
}

// printHeader displays a nice visual header with the tool name and version
func printHeader() {
	// Use bold cyan for the main text
	headerColor := color.New(color.FgCyan, color.Bold)
	versionColor := color.New(color.FgHiBlack)

	// Print to stderr - this is status/diagnostic info, not the actual output
	headerColor.Fprint(os.Stderr, "[ Que? ]")
	versionColor.Fprintf(os.Stderr, "  version %s\n\n", Version)
}
