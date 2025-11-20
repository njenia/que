package enricher

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jenian/que/internal/config"
)

// Enrich gathers system context information
func Enrich() config.Context {
	ctx := config.Context{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now(),
	}

	// Get shell from environment variable
	if shell := os.Getenv("SHELL"); shell != "" {
		ctx.Shell = filepath.Base(shell)
	} else {
		// Try to detect shell from ps command
		ctx.Shell = detectShell()
	}

	return ctx
}

// detectShell attempts to detect the current shell by checking the parent process
func detectShell() string {
	// Try common shell detection methods
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}

	// Try to get shell from ps command (Unix-like systems)
	if runtime.GOOS != "windows" {
		cmd := exec.Command("ps", "-p", "$$", "-o", "comm=")
		output, err := cmd.Output()
		if err == nil {
			shell := strings.TrimSpace(string(output))
			if shell != "" {
				return shell
			}
		}
	}

	// Fallback
	return "unknown"
}

