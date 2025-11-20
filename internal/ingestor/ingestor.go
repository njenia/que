package ingestor

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

const (
	// MaxInputSize is the maximum size of input before truncation (100KB)
	MaxInputSize = 100 * 1024
	// TruncateHeadSize is the size of the head to keep when truncating (50KB)
	TruncateHeadSize = 50 * 1024
	// TruncateTailSize is the size of the tail to keep when truncating (50KB)
	TruncateTailSize = 50 * 1024
)

// Ingest reads from stdin and returns the content, with intelligent truncation
// if the input exceeds MaxInputSize. When truncating, it preserves the head
// and tail of the input while maintaining line boundaries.
func Ingest() (string, error) {
	return IngestFromReader(os.Stdin)
}

// IngestFromReader reads from the provided reader and returns the content
func IngestFromReader(r io.Reader) (string, error) {
	reader := bufio.NewReader(r)
	var buffer bytes.Buffer
	
	// Read all input
	_, err := buffer.ReadFrom(reader)
	if err != nil && err != io.EOF {
		return "", err
	}
	
	content := buffer.Bytes()
	
	// If content is within limits, return as-is
	if len(content) <= MaxInputSize {
		return string(content), nil
	}
	
	// Truncate: keep head + tail
	head := content[:TruncateHeadSize]
	tail := content[len(content)-TruncateTailSize:]
	
	// Find the last newline in the head to preserve line boundaries
	headLastNewline := bytes.LastIndexByte(head, '\n')
	if headLastNewline == -1 {
		// No newline found, use full head
		headLastNewline = len(head)
	} else {
		headLastNewline++ // Include the newline
	}
	
	// Find the first newline in the tail to preserve line boundaries
	tailFirstNewline := bytes.IndexByte(tail, '\n')
	if tailFirstNewline == -1 {
		// No newline found, use full tail
		tailFirstNewline = 0
	} else {
		tailFirstNewline++ // Include the newline
	}
	
	// Combine head and tail with truncation indicator
	truncated := bytes.NewBuffer(head[:headLastNewline])
	truncated.WriteString("\n... [TRUNCATED: input exceeded 100KB, showing first 50KB and last 50KB] ...\n")
	truncated.Write(tail[tailFirstNewline:])
	
	return truncated.String(), nil
}

