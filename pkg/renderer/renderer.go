package renderer

import (
	"bytes"
	"os"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown content for terminal output
func RenderMarkdown(content string) (string, error) {
	// Create a glamour renderer with terminal-appropriate settings
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return "", err
	}

	// Render the markdown
	out, err := r.Render(content)
	if err != nil {
		return "", err
	}

	return out, nil
}

// RenderToStdout renders markdown and writes directly to stdout
func RenderToStdout(content string) error {
	rendered, err := RenderMarkdown(content)
	if err != nil {
		return err
	}

	_, err = os.Stdout.WriteString(rendered)
	return err
}

// RenderToBuffer renders markdown to a buffer (useful for testing)
func RenderToBuffer(content string) (*bytes.Buffer, error) {
	rendered, err := RenderMarkdown(content)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(rendered)
	return buf, nil
}

