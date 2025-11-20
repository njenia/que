package renderer

import (
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	markdown := "# Test\n\nThis is a **test** markdown."
	
	result, err := RenderMarkdown(markdown)
	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v, want nil", err)
	}
	
	if result == "" {
		t.Error("RenderMarkdown() should return non-empty result")
	}
	
	// Should contain some rendered content (exact format depends on terminal)
	if len(result) < len(markdown) {
		t.Error("Rendered markdown should be at least as long as input")
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	markdown := "```bash\necho 'hello'\n```"
	
	result, err := RenderMarkdown(markdown)
	if err != nil {
		t.Fatalf("RenderMarkdown() error = %v, want nil", err)
	}
	
	if result == "" {
		t.Error("RenderMarkdown() should return non-empty result")
	}
}

