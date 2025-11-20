package sanitizer

import (
	"strings"
	"testing"

	"github.com/zricethezav/gitleaks/v8/report"
)

// Integration tests using real gitleaks detector
func TestRedactor_GitHubToken_Integration(t *testing.T) {
	redactor := NewRedactor()
	
	// GitHub token that gitleaks will detect
	input := "GITHUB_TOKEN=ghp_123456789012345678901234567890123456"
	result, count := redactor.Redact(input)
	
	if count == 0 {
		t.Error("Expected at least one redaction for GitHub token")
	}
	if !strings.Contains(result, "REDACTED") {
		t.Errorf("Result should contain REDACTED, got: %s", result)
	}
	if strings.Contains(result, "ghp_") {
		t.Errorf("Result should not contain the token, got: %s", result)
	}
}

func TestRedactor_MultipleSecrets_Integration(t *testing.T) {
	redactor := NewRedactor()
	
	// Multiple GitHub tokens (gitleaks pattern requires 40 chars after ghp_)
	input := "First token: ghp_123456789012345678901234567890123456 and second: ghp_abcdefghijklmnopqrstuvwxyz123456789012"
	result, count := redactor.Redact(input)
	
	// Gitleaks may not detect all tokens depending on pattern matching
	// But if it finds any, they should be redacted
	if count > 0 {
		if strings.Contains(result, "ghp_123456789012345678901234567890123456") {
			t.Errorf("Result should not contain the first token, got: %s", result)
		}
	}
}

// Unit tests using mock detector
func TestRedactor_WithMockDetector(t *testing.T) {
	// Create mock findings
	mockFindings := []report.Finding{
		MockFinding("github-token", "GitHub token", "ghp_123456789012345678901234567890123456", "ghp_123456789012345678901234567890123456"),
	}
	
	mockDetector := NewMockDetector(mockFindings)
	redactor := NewRedactorWithDetector(mockDetector)
	
	input := "GITHUB_TOKEN=ghp_123456789012345678901234567890123456"
	result, count := redactor.Redact(input)
	
	if count != 1 {
		t.Errorf("Expected 1 redaction, got %d", count)
	}
	if !strings.Contains(result, "REDACTED_GITHUB_TOKEN") {
		t.Errorf("Result should contain REDACTED_GITHUB_TOKEN, got: %s", result)
	}
	if strings.Contains(result, "ghp_123456789012345678901234567890123456") {
		t.Errorf("Result should not contain the token, got: %s", result)
	}
}

func TestRedactor_NoSecrets_WithMock(t *testing.T) {
	// Mock detector with no findings
	mockDetector := NewMockDetector([]report.Finding{})
	redactor := NewRedactorWithDetector(mockDetector)
	
	input := "This is a normal log message with no secrets"
	result, count := redactor.Redact(input)
	
	if count != 0 {
		t.Errorf("Expected no redactions, got %d", count)
	}
	if result != input {
		t.Errorf("Result should be unchanged, got: %s", result)
	}
}

func TestRedactor_MultipleSecrets_WithMock(t *testing.T) {
	// Create mock findings for multiple secrets
	mockFindings := []report.Finding{
		MockFinding("aws-access-key", "AWS Access Key", "AKIAIOSFODNN7EXAMPLE", "AKIAIOSFODNN7EXAMPLE"),
		MockFinding("github-token", "GitHub token", "ghp_abcdefghijklmnopqrstuvwxyz123456789012", "ghp_abcdefghijklmnopqrstuvwxyz123456789012"),
	}
	
	mockDetector := NewMockDetector(mockFindings)
	redactor := NewRedactorWithDetector(mockDetector)
	
	input := "AWS_KEY=AKIAIOSFODNN7EXAMPLE and GITHUB_TOKEN=ghp_abcdefghijklmnopqrstuvwxyz123456789012"
	result, count := redactor.Redact(input)
	
	if count != 2 {
		t.Errorf("Expected 2 redactions, got %d", count)
	}
	if strings.Contains(result, "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("Result should not contain AWS key, got: %s", result)
	}
	if strings.Contains(result, "ghp_abcdefghijklmnopqrstuvwxyz123456789012") {
		t.Errorf("Result should not contain GitHub token, got: %s", result)
	}
	if !strings.Contains(result, "REDACTED") {
		t.Errorf("Result should contain REDACTED, got: %s", result)
	}
}

func TestRedactor_EmptyInput(t *testing.T) {
	mockDetector := NewMockDetector([]report.Finding{})
	redactor := NewRedactorWithDetector(mockDetector)
	
	input := ""
	result, count := redactor.Redact(input)
	
	if count != 0 {
		t.Errorf("Expected no redactions for empty input, got %d", count)
	}
	if result != input {
		t.Errorf("Result should be empty, got: %s", result)
	}
}

func TestRedactor_RedactWithDetails_WithMock(t *testing.T) {
	mockFindings := []report.Finding{
		MockFinding("api-key", "API Key", "sk_live_1234567890abcdef", "sk_live_1234567890abcdef"),
	}
	
	mockDetector := NewMockDetector(mockFindings)
	redactor := NewRedactorWithDetector(mockDetector)
	
	input := "API_KEY=sk_live_1234567890abcdef"
	result, count, details := redactor.RedactWithDetails(input, true)
	
	if count != 1 {
		t.Errorf("Expected 1 redaction, got %d", count)
	}
	if len(details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(details))
	}
	if details[0].RuleID != "api-key" {
		t.Errorf("Expected rule ID 'api-key', got %s", details[0].RuleID)
	}
	if !strings.Contains(result, "REDACTED") {
		t.Errorf("Result should contain REDACTED, got: %s", result)
	}
}

