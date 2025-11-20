package sanitizer

import (
	"os"
	"strings"

	"github.com/jenian/que/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	gitleaksconfig "github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/report"
)

// Detector is an interface for secret detection
type Detector interface {
	DetectString(content string) []report.Finding
}

// redactor implements the Redactor interface using gitleaks for detection
type redactor struct {
	detector Detector
}

// NewRedactor creates a new redactor with gitleaks detector and custom rules
func NewRedactor() config.Redactor {
	return NewRedactorWithDetector(nil)
}

// NewRedactorWithDetector creates a new redactor with the provided detector.
// If detector is nil, it creates a default gitleaks detector.
func NewRedactorWithDetector(detector Detector) config.Redactor {
	// Disable gitleaks logging
	zerolog.SetGlobalLevel(zerolog.Disabled)

	// If detector is provided, use it
	if detector != nil {
		return &redactor{detector: detector}
	}

	// Otherwise, create default gitleaks detector
	cfg, err := loadDefaultConfig()
	if err != nil {
		// Fallback to empty config if we can't load default
		defaultDetector := detect.NewDetector(gitleaksconfig.Config{})
		return &redactor{detector: defaultDetector}
	}

	defaultDetector := detect.NewDetector(cfg)
	return &redactor{
		detector: defaultDetector,
	}
}

// loadDefaultConfig loads the default gitleaks configuration and merges custom rules
func loadDefaultConfig() (gitleaksconfig.Config, error) {
	// Use the same approach as NewDetectorDefaultConfig
	// Reset viper to avoid state issues
	viper.Reset()
	viper.SetConfigType("toml")

	// Load default config
	err := viper.ReadConfig(strings.NewReader(gitleaksconfig.DefaultConfig))
	if err != nil {
		return gitleaksconfig.Config{}, err
	}

	// Try to load custom rules from TOML file
	customConfigPath := ".gitleaks-custom.toml"
	if _, err := os.Stat(customConfigPath); err == nil {
		// Custom config file exists, merge it
		viper.SetConfigFile(customConfigPath)
		err = viper.MergeInConfig()
		if err != nil {
			// If merge fails, continue with default config only
			// This is not a fatal error
		}
	}

	var vc gitleaksconfig.ViperConfig
	err = viper.Unmarshal(&vc)
	if err != nil {
		return gitleaksconfig.Config{}, err
	}

	cfg, err := vc.Translate()
	if err != nil {
		return gitleaksconfig.Config{}, err
	}

	return cfg, nil
}

// Redact uses gitleaks to detect secrets and redacts them from the input
func (r *redactor) Redact(input string) (string, int) {
	result, count, _ := r.RedactWithDetails(input, false)
	return result, count
}

// RedactWithDetails uses gitleaks to detect secrets and redacts them from the input
// Returns redacted string, count, and details for verbose output
func (r *redactor) RedactWithDetails(input string, verbose bool) (string, int, []config.FindingDetail) {
	if input == "" {
		return input, 0, nil
	}

	result := input
	totalRedactions := 0
	var details []config.FindingDetail

	// Use gitleaks to detect secrets
	findings := r.detector.DetectString(input)

	// Track redactions to avoid double-redacting
	redacted := make(map[string]bool)

	// Process gitleaks findings
	if len(findings) > 0 {
		// First, collect all findings for verbose output and determine unique secrets to redact
		uniqueSecrets := make(map[string]string) // secret -> placeholder

		for _, finding := range findings {
			// Use the Match field which contains the matched text
			secret := finding.Match
			if secret == "" {
				// Fallback to Secret field if Match is empty
				secret = finding.Secret
			}

			if secret == "" {
				continue
			}

			// Always add to details for verbose output (even if same secret appears multiple times)
			if verbose {
				details = append(details, config.FindingDetail{
					RuleID:      finding.RuleID,
					Description: finding.Description,
					Match:       finding.Match,
					Secret:      finding.Secret,
					Source:      "gitleaks",
					StartLine:   finding.StartLine,
					EndLine:     finding.EndLine,
					StartColumn: finding.StartColumn,
					EndColumn:   finding.EndColumn,
					Entropy:     finding.Entropy,
					File:        finding.File,
					Tags:        finding.Tags,
					Fingerprint: finding.Fingerprint,
				})
			}

			// Track unique secrets and their placeholders (only need to redact once per unique secret)
			if _, exists := uniqueSecrets[secret]; !exists {
				placeholder := getRedactionPlaceholder(finding.RuleID, finding.Description)
				uniqueSecrets[secret] = placeholder
			}
		}

		// Now redact all unique secrets
		for secret, placeholder := range uniqueSecrets {
			if !redacted[secret] {
				count := strings.Count(result, secret)
				if count > 0 {
					result = strings.ReplaceAll(result, secret, placeholder)
					redacted[secret] = true
					totalRedactions += count
				}
			}
		}
	}

	return result, totalRedactions, details
}

// getRedactionPlaceholder returns an appropriate redaction placeholder based on the rule
func getRedactionPlaceholder(ruleID, description string) string {
	// Map common rule IDs to specific placeholders
	ruleIDLower := strings.ToLower(ruleID)
	descLower := strings.ToLower(description)

	// Check rule ID first
	if strings.Contains(ruleIDLower, "aws") {
		if strings.Contains(ruleIDLower, "access") || strings.Contains(descLower, "access key") {
			return "<REDACTED_AWS_ACCESS_KEY>"
		}
		if strings.Contains(ruleIDLower, "secret") || strings.Contains(descLower, "secret") {
			return "<REDACTED_AWS_SECRET_KEY>"
		}
		return "<REDACTED_AWS_CREDENTIAL>"
	}

	if strings.Contains(ruleIDLower, "github") || strings.Contains(descLower, "github") {
		return "<REDACTED_GITHUB_TOKEN>"
	}

	if strings.Contains(ruleIDLower, "private") || strings.Contains(descLower, "private key") {
		return "<REDACTED_PRIVATE_KEY>"
	}

	if strings.Contains(ruleIDLower, "api") || strings.Contains(descLower, "api key") {
		return "<REDACTED_API_KEY>"
	}

	if strings.Contains(ruleIDLower, "bearer") || strings.Contains(descLower, "bearer") {
		return "<REDACTED_BEARER_TOKEN>"
	}

	if strings.Contains(ruleIDLower, "token") || strings.Contains(descLower, "token") {
		return "<REDACTED_TOKEN>"
	}

	if strings.Contains(ruleIDLower, "password") || strings.Contains(descLower, "password") {
		return "<REDACTED_PASSWORD>"
	}

	if strings.Contains(ruleIDLower, "db-connection") || strings.Contains(descLower, "database connection") {
		return "<REDACTED_DB_CONNECTION_STRING>"
	}

	// Generic placeholder for unknown secret types
	return "<REDACTED_SECRET>"
}
