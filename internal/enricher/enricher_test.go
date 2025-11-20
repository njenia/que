package enricher

import (
	"testing"
)

func TestEnrich(t *testing.T) {
	ctx := Enrich()
	
	if ctx.OS == "" {
		t.Error("Context OS should not be empty")
	}
	
	if ctx.Arch == "" {
		t.Error("Context Arch should not be empty")
	}
	
	if ctx.Timestamp.IsZero() {
		t.Error("Context Timestamp should not be zero")
	}
	
	// Shell might be empty if detection fails, but that's okay
	// We just check that the function doesn't panic
}

