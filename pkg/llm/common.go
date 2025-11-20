package llm

import (
	"fmt"
	"strings"
	"time"

	"github.com/jenian/que/internal/config"
)

// formatPrompt formats the payload into a user-friendly prompt for the LLM
func formatPrompt(payload config.QueryPayload) string {
	var parts []string

	// Add system context if available
	if payload.SystemContext.OS != "" {
		ctx := payload.SystemContext
		contextInfo := fmt.Sprintf("System Environment:\n- OS: %s\n- Architecture: %s\n- Shell: %s\n- Timestamp: %s",
			ctx.OS, ctx.Arch, ctx.Shell, ctx.Timestamp.Format(time.RFC3339))
		parts = append(parts, contextInfo)
	}

	// Add the sanitized log
	parts = append(parts, "Log/Error Data:")
	parts = append(parts, payload.SanitizedLog)

	// Add instruction
	parts = append(parts, "\nAnalyze the above log data and return a strict JSON response with exactly four fields:")
	parts = append(parts, "1. \"status\": One of: \"no_problem\" (if no errors/issues detected), \"insufficient_data\" (if problem detected but not enough info for a clear solution), or \"problem_detected\" (if problem found with clear solution)")
	parts = append(parts, "2. \"root_cause\": A concise description of the root cause (empty string if status is \"no_problem\")")
	parts = append(parts, "3. \"evidence\": The relevant log lines that indicate the problem (include actual log lines, or empty string if status is \"no_problem\")")
	parts = append(parts, "4. \"fix\": A concise, executable CLI fix or code patch (empty string if status is \"no_problem\" or \"insufficient_data\")")
	parts = append(parts, "\nReturn ONLY valid JSON, no markdown, no code blocks, no explanations. The JSON must be parseable.")

	return strings.Join(parts, "\n\n")
}
