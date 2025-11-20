package advisor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/jenian/que/internal/config"
	"github.com/jenian/que/pkg/llm"
)

// Advise processes the payload and returns formatted advice from the LLM
func Advise(client llm.Client, cfg *config.Config, payload config.QueryPayload) (string, error) {
	// Handle dry-run mode
	if cfg.DryRun {
		return handleDryRun(cfg, payload)
	}

	// Show spinner while waiting for LLM response
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Analyzing..."
	s.Writer = os.Stderr
	s.Start()
	defer s.Stop() // Always stop spinner, even on error

	// Query the LLM using the injected client
	response, err := client.QueryWithPayload(cfg, payload)
	if err != nil {
		return "", err
	}

	// Parse and format the JSON response
	formatted, err := parseAndFormatResponse(response)
	if err != nil {
		// If parsing fails, return the raw response with an error message
		return fmt.Sprintf("Error parsing LLM response: %v\n\nRaw response:\n%s", err, response), nil
	}

	return formatted, nil
}

// handleDryRun shows what would be sent without making an API call
func handleDryRun(cfg *config.Config, payload config.QueryPayload) (string, error) {
	var output string

	output += "=== DRY RUN MODE ===\n\n"
	output += fmt.Sprintf("Provider: %s\n", cfg.Provider)
	if cfg.Model != "" {
		output += fmt.Sprintf("Model: %s\n", cfg.Model)
	}
	output += "\n"

	// Show system context
	if payload.SystemContext.OS != "" {
		output += "System Context:\n"
		output += fmt.Sprintf("  OS: %s\n", payload.SystemContext.OS)
		output += fmt.Sprintf("  Arch: %s\n", payload.SystemContext.Arch)
		output += fmt.Sprintf("  Shell: %s\n", payload.SystemContext.Shell)
		output += fmt.Sprintf("  Timestamp: %s\n", payload.SystemContext.Timestamp.Format("2006-01-02 15:04:05"))
		output += "\n"
	}

	// Show sanitized log preview
	output += "Sanitized Log (first 500 chars):\n"
	preview := payload.SanitizedLog
	if len(preview) > 500 {
		preview = preview[:500] + "... [truncated]"
	}
	output += preview + "\n\n"

	output += "Full sanitized log length: " + fmt.Sprintf("%d", len(payload.SanitizedLog)) + " characters\n"
	output += "\n"
	output += "=== Would query LLM API (skipped in dry-run mode) ===\n"

	// Write to stderr so it doesn't interfere with pipeline usage
	fmt.Fprint(os.Stderr, output)
	return "", nil
}

// extractJSON extracts JSON from a response that might be wrapped in markdown code blocks
func extractJSON(response string) string {
	// Try to find JSON in markdown code blocks
	jsonBlockRegex := regexp.MustCompile(`(?s)` + "```" + `(?:json)?\s*\n(.*?)\n` + "```")
	if matches := jsonBlockRegex.FindStringSubmatch(response); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find JSON object directly
	jsonObjRegex := regexp.MustCompile(`(?s)\{(.*)\}`)
	if matches := jsonObjRegex.FindStringSubmatch(response); len(matches) > 0 {
		return strings.TrimSpace(matches[0])
	}

	// Return trimmed response as-is
	return strings.TrimSpace(response)
}

// parseAndFormatResponse parses the JSON response and formats it for console output
func parseAndFormatResponse(rawResponse string) (string, error) {
	// Extract JSON from potential markdown wrappers
	jsonStr := extractJSON(rawResponse)

	// Parse JSON
	var llmResp config.LLMResponse
	if err := json.Unmarshal([]byte(jsonStr), &llmResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Handle different status cases
	status := strings.ToLower(strings.TrimSpace(llmResp.Status))

	// Case 1: no problems detected
	if status == "no_problem" || (status == "" && strings.TrimSpace(llmResp.RootCause) == "" && strings.TrimSpace(string(llmResp.Evidence)) == "") {
		return "Your log looks good, no problems detected!\n", nil
	}

	// Case 2: Problem detected but insufficient data
	if status == "insufficient_data" || (status == "" && strings.TrimSpace(llmResp.Fix) == "" && strings.TrimSpace(string(llmResp.Evidence)) != "") {
		var output strings.Builder
		titleColor := color.New(color.FgCyan, color.Bold)
		messageColor := color.New(color.FgYellow)

		// Show evidence
		output.WriteString(titleColor.Sprint("Evidence"))
		output.WriteString("\n\n")
		evidenceLines := strings.Split(strings.TrimSpace(string(llmResp.Evidence)), "\n")
		for _, line := range evidenceLines {
			output.WriteString(line)
			output.WriteString("\n")
		}
		output.WriteString("\n")

		// Show message
		output.WriteString(messageColor.Sprint("⚠️  Problem detected but insufficient data for a clear solution. Please provide more context or logs."))
		output.WriteString("\n")

		return output.String(), nil
	}

	// Case 3: Problem detected with solution - show full output
	titleColor := color.New(color.FgCyan, color.Bold)
	var output strings.Builder

	// Root Cause section
	if strings.TrimSpace(llmResp.RootCause) != "" {
		output.WriteString(titleColor.Sprintln())
		output.WriteString(titleColor.Sprint("Root Cause"))
		output.WriteString("\n\n")
		output.WriteString(strings.TrimSpace(llmResp.RootCause))
		output.WriteString("\n\n")
	}

	// Evidence section
	if strings.TrimSpace(string(llmResp.Evidence)) != "" {
		output.WriteString(titleColor.Sprint("Evidence"))
		output.WriteString("\n\n")
		evidenceLines := strings.Split(strings.TrimSpace(string(llmResp.Evidence)), "\n")
		for _, line := range evidenceLines {
			output.WriteString(line)
			output.WriteString("\n")
		}
		output.WriteString("\n")
	}

	// Fix section
	if strings.TrimSpace(llmResp.Fix) != "" {
		output.WriteString(titleColor.Sprint("Fix"))
		output.WriteString("\n\n")
		fixLines := strings.Split(strings.TrimSpace(llmResp.Fix), "\n")
		for _, line := range fixLines {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	return output.String(), nil
}

// RunInteractive starts an interactive conversation session
func RunInteractive(client llm.Client, cfg *config.Config, payload config.QueryPayload, initialResponse string) error {
	// Build initial user message with log context
	initialUserMessage := buildInitialUserMessage(payload)

	// Conversation history: [user1, assistant1, user2, assistant2, ...]
	conversationHistory := []string{
		initialUserMessage, // User: "Here's the log, analyze it"
		initialResponse,    // Assistant: Initial analysis
	}

	// Create a prompt color for better UX
	promptColor := color.New(color.FgCyan, color.Bold)

	fmt.Fprintf(os.Stderr, "\n")
	promptColor.Fprintf(os.Stderr, "💬 Interactive mode - Ask follow-up questions (type 'exit' or 'quit' to exit)\n\n")

	// Open terminal for reading (works even when stdin is piped)
	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		// Fallback to stdin if /dev/tty is not available (e.g., Windows)
		tty = os.Stdin
	} else {
		defer tty.Close()
	}

	scanner := bufio.NewScanner(tty)

	for {
		// Prompt for user input
		promptColor.Fprintf(os.Stderr, "> ")

		if !scanner.Scan() {
			// EOF or error
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		// Check for exit commands
		if userInput == "" {
			continue
		}
		if userInput == "exit" || userInput == "quit" || userInput == "q" {
			fmt.Fprintf(os.Stderr, "Exiting interactive mode.\n")
			break
		}

		// Show spinner while waiting for response
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Thinking..."
		s.Writer = os.Stderr
		s.Start()

		// Query LLM with follow-up question using the injected client
		response, err := client.QueryWithHistory(cfg, conversationHistory, userInput)

		s.Stop()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		// Display response
		fmt.Print(response)
		fmt.Print("\n\n")

		// Update conversation history
		conversationHistory = append(conversationHistory, userInput, response)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// buildInitialUserMessage creates the initial user message with log context
func buildInitialUserMessage(payload config.QueryPayload) string {
	var parts []string

	parts = append(parts, "Please analyze the following log data:")
	parts = append(parts, "")

	// Include system context if available (same format as initial query)
	if payload.SystemContext.OS != "" {
		ctx := payload.SystemContext
		contextInfo := fmt.Sprintf("System Environment:\n- OS: %s\n- Architecture: %s\n- Shell: %s\n- Timestamp: %s",
			ctx.OS, ctx.Arch, ctx.Shell, ctx.Timestamp.Format(time.RFC3339))
		parts = append(parts, contextInfo)
		parts = append(parts, "")
	}

	parts = append(parts, "=== Original Log Data ===")

	// Include sanitized log (first 2000 chars to keep context manageable)
	logPreview := payload.SanitizedLog
	if len(logPreview) > 2000 {
		logPreview = logPreview[:2000] + "... [truncated]"
	}
	parts = append(parts, logPreview)
	parts = append(parts, "")
	parts = append(parts, "=== End Original Log Data ===")

	return strings.Join(parts, "\n")
}
