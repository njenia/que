package advisor

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jenian/que/internal/config"
)

// mockLLMResponse creates a mock LLM JSON response
func mockLLMResponse(status, rootCause, evidence, fix string) string {
	resp := config.LLMResponse{
		Status:    status,
		RootCause: rootCause,
		Evidence:  config.EvidenceString(evidence),
		Fix:       fix,
	}
	jsonData, _ := json.Marshal(resp)
	return string(jsonData)
}

func TestParseAndFormatResponse_NoProblem(t *testing.T) {
	// Test case 1: no problems detected
	mockResponse := mockLLMResponse("no_problem", "", "", "")
	
	result, err := parseAndFormatResponse(mockResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Your log looks good, no problems detected!\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestParseAndFormatResponse_NoProblemFallback(t *testing.T) {
	// Test case 1b: no problems detected (fallback - empty fields)
	mockResponse := mockLLMResponse("", "", "", "")
	
	result, err := parseAndFormatResponse(mockResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expected := "Your log looks good, no problems detected!\n"
	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestParseAndFormatResponse_InsufficientData(t *testing.T) {
	// Test case 2: Problem detected but insufficient data
	evidence := "2024-01-15 14:32:11 ERROR Something went wrong\nStatus: 500 Internal Server Error"
	mockResponse := mockLLMResponse("insufficient_data", "Server error occurred", evidence, "")
	
	result, err := parseAndFormatResponse(mockResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should contain evidence
	if !strings.Contains(result, "Evidence") {
		t.Error("Result should contain 'Evidence' section")
	}
	if !strings.Contains(result, evidence) {
		t.Error("Result should contain the evidence text")
	}
	
	// Should contain warning message
	if !strings.Contains(result, "insufficient data") {
		t.Error("Result should contain insufficient data warning")
	}
	
	// Should NOT contain fix
	if strings.Contains(result, "Fix") {
		t.Error("Result should NOT contain 'Fix' section for insufficient data")
	}
}

func TestParseAndFormatResponse_InsufficientDataFallback(t *testing.T) {
	// Test case 2b: Insufficient data (fallback - evidence but no fix)
	evidence := "2024-01-15 14:32:11 ERROR Something went wrong"
	mockResponse := mockLLMResponse("", "Some error", evidence, "")
	
	result, err := parseAndFormatResponse(mockResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should contain evidence
	if !strings.Contains(result, "Evidence") {
		t.Error("Result should contain 'Evidence' section")
	}
	
	// Should contain warning message
	if !strings.Contains(result, "insufficient data") {
		t.Error("Result should contain insufficient data warning")
	}
}

func TestParseAndFormatResponse_ProblemWithSolution(t *testing.T) {
	// Test case 3: Problem detected with clear solution
	rootCause := "Invalid API key"
	evidence := "2024-01-15 14:32:11 ERROR Authentication failed\nStatus: 401 Unauthorized"
	fix := "export API_KEY='your_valid_key'"
	mockResponse := mockLLMResponse("problem_detected", rootCause, evidence, fix)
	
	result, err := parseAndFormatResponse(mockResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should contain all three sections
	if !strings.Contains(result, "Root Cause") {
		t.Error("Result should contain 'Root Cause' section")
	}
	if !strings.Contains(result, rootCause) {
		t.Error("Result should contain root cause text")
	}
	
	if !strings.Contains(result, "Evidence") {
		t.Error("Result should contain 'Evidence' section")
	}
	if !strings.Contains(result, evidence) {
		t.Error("Result should contain evidence text")
	}
	
	if !strings.Contains(result, "Fix") {
		t.Error("Result should contain 'Fix' section")
	}
	if !strings.Contains(result, fix) {
		t.Error("Result should contain fix text")
	}
}

func TestParseAndFormatResponse_EvidenceAsArray(t *testing.T) {
	// Test evidence as array (should be handled by EvidenceString unmarshaler)
	// Create response with evidence as array
	resp := map[string]interface{}{
		"status":     "problem_detected",
		"root_cause": "Server error",
		"evidence":   []string{"2024-01-15 14:32:11 ERROR", "Status: 500", "Error details"},
		"fix":        "Check server logs",
	}
	jsonData, _ := json.Marshal(resp)
	
	result, err := parseAndFormatResponse(string(jsonData))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should contain evidence lines
	if !strings.Contains(result, "2024-01-15 14:32:11 ERROR") {
		t.Error("Result should contain first evidence line")
	}
	if !strings.Contains(result, "Status: 500") {
		t.Error("Result should contain second evidence line")
	}
}

func TestParseAndFormatResponse_JSONInMarkdown(t *testing.T) {
	// Test JSON wrapped in markdown code blocks
	jsonResponse := mockLLMResponse("problem_detected", "Test root cause", "Test evidence", "Test fix")
	wrappedResponse := "```json\n" + jsonResponse + "\n```"
	
	result, err := parseAndFormatResponse(wrappedResponse)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Should parse correctly
	if !strings.Contains(result, "Root Cause") {
		t.Error("Should parse JSON from markdown code block")
	}
}

func TestAdvise_DryRun(t *testing.T) {
	cfg := &config.Config{
		DryRun:   true,
		Provider: "openai",
		Verbose:  false,
	}
	
	payload := config.QueryPayload{
		RawLog:       "Test log message",
		SanitizedLog: "Test log message",
		SystemContext: config.Context{
			OS:        "linux",
			Arch:      "amd64",
			Shell:     "bash",
			Timestamp: config.Context{}.Timestamp, // Will be zero time
		},
	}
	
	// In dry-run mode, client can be nil
	result, err := Advise(nil, cfg, payload)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Dry run should return empty string (output goes to stderr)
	if result != "" {
		t.Errorf("Expected empty string for dry run, got: %s", result)
	}
}

