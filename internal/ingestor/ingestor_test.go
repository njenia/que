package ingestor

import (
	"strings"
	"testing"
)

func TestIngestFromReader_SmallInput(t *testing.T) {
	// Test with small input that doesn't need truncation
	input := "This is a small log file\nwith multiple lines\nof content."
	reader := strings.NewReader(input)
	
	result, err := IngestFromReader(reader)
	if err != nil {
		t.Fatalf("IngestFromReader() error = %v, want nil", err)
	}
	
	if result != input {
		t.Errorf("IngestFromReader() = %q, want %q", result, input)
	}
}

func TestIngestFromReader_LargeInput(t *testing.T) {
	// Test with large input that needs truncation
	head := strings.Repeat("A", TruncateHeadSize)
	tail := strings.Repeat("B", TruncateTailSize)
	middle := strings.Repeat("C", 100*1024) // 100KB of middle content
	input := head + "\n" + middle + "\n" + tail
	reader := strings.NewReader(input)
	
	result, err := IngestFromReader(reader)
	if err != nil {
		t.Fatalf("IngestFromReader() error = %v, want nil", err)
	}
	
	// Should contain truncation message
	if !strings.Contains(result, "[TRUNCATED") {
		t.Error("IngestFromReader() should contain truncation message for large input")
	}
	
	// Should contain head
	if !strings.Contains(result, "A") {
		t.Error("IngestFromReader() should contain head portion")
	}
	
	// Should contain tail
	if !strings.Contains(result, "B") {
		t.Error("IngestFromReader() should contain tail portion")
	}
	
	// Should not contain middle
	if strings.Contains(result, "CCCCCC") {
		t.Error("IngestFromReader() should not contain middle portion")
	}
}

func TestIngestFromReader_EmptyInput(t *testing.T) {
	reader := strings.NewReader("")
	
	result, err := IngestFromReader(reader)
	if err != nil {
		t.Fatalf("IngestFromReader() error = %v, want nil", err)
	}
	
	if result != "" {
		t.Errorf("IngestFromReader() = %q, want empty string", result)
	}
}

