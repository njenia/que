package sanitizer

import (
	"github.com/zricethezav/gitleaks/v8/report"
)

// MockDetector is a mock implementation of the Detector interface for testing
type MockDetector struct {
	findings []report.Finding
}

// NewMockDetector creates a new mock detector with the provided findings
func NewMockDetector(findings []report.Finding) *MockDetector {
	return &MockDetector{
		findings: findings,
	}
}

// DetectString returns the mock findings
func (m *MockDetector) DetectString(content string) []report.Finding {
	return m.findings
}

// MockFinding creates a mock finding for testing
func MockFinding(ruleID, description, match, secret string) report.Finding {
	return report.Finding{
		RuleID:      ruleID,
		Description: description,
		Match:       match,
		Secret:      secret,
		StartLine:   1,
		EndLine:     1,
		StartColumn: 0,
		EndColumn:   len(match),
		Entropy:     4.5,
		Tags:        []string{"test"},
	}
}

