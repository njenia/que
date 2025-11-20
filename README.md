# Que
> The pipe-able DevOps assistant.

  <img src="img/que_demo.gif" alt="Demo" style="flex: 1" />

Que is a CLI utility designed to act as a filter in a Unix pipeline. It ingests stdin (logs, error tracebacks, config files), sanitizes the data for security, enriches it with local system context, and queries an LLM (ChatGPT or Claude) to determine the root cause and suggest a fix. Perfect for analyzing errors in servers and CI/CD environments where you don't have easy access to AI-powered editors.

## 🔒 Privacy & Security

**Que runs entirely locally.** It scrubs secrets (API keys, PII) using Gitleaks rules before the request leaves your machine. Logs are stateless and not stored.



## Installation

### Quick Install

**Universal installer (auto-detects platform):**
```bash
curl -sSL https://raw.githubusercontent.com/njenia/que/main/install.sh | bash
```

Or download and install the latest release for your platform:

**Linux (amd64):**
```bash
curl -L https://github.com/njenia/que/releases/latest/download/que-linux-amd64.tar.gz | tar -xz && sudo mv que /usr/local/bin/
```

**Linux (arm64):**
```bash
curl -L https://github.com/njenia/que/releases/latest/download/que-linux-arm64.tar.gz | tar -xz && sudo mv que /usr/local/bin/
```

**macOS (amd64):**
```bash
curl -L https://github.com/njenia/que/releases/latest/download/que-darwin-amd64.tar.gz | tar -xz && sudo mv que /usr/local/bin/
```

**macOS (arm64 / Apple Silicon):**
```bash
curl -L https://github.com/njenia/que/releases/latest/download/que-darwin-arm64.tar.gz | tar -xz && sudo mv que /usr/local/bin/
```

**Windows:**
```powershell
# Download and extract
curl -L https://github.com/njenia/que/releases/latest/download/que-windows-amd64.zip -o que.zip
Expand-Archive que.zip
# Move que.exe to a directory in your PATH
```

### From Source

```bash
# Clone the repository
git clone https://github.com/njenia/que.git
cd que

# Build with Make (automatically detects version from git tags)
make build

# Or install directly
make install
```

### Using Go Install

```bash
go install github.com/njenia/que/cmd/que@latest
```

**Note**: When building locally, use `make build` to automatically set the version from git tags. Building with `go build` directly will show version as "dev".

## Usage

### Basic Usage

```bash
# Defaults to ChatGPT
cat server.log | que

# With specific provider
tail -n 50 error.log | que --provider claude

# Strict mode (verbose output showing what's being sent)
tail -n 50 error.log | que --provider claude --verbose
```

### Configuration

Set your API keys as environment variables:

```bash
export QUE_CHATGPT_API_KEY="your-openai-api-key"
export QUE_CLAUDE_API_KEY="your-anthropic-api-key"
export QUE_DEFAULT_PROVIDER="openai"  # Optional, defaults to openai
```

### CLI Flags

- `-p, --provider string`: LLM provider to use (openai, claude)
- `-m, --model string`: Specific model override (e.g., gpt-4-turbo)
- `-v, --verbose`: Show what data is being sent (including redaction)
- `-i, --interactive`: Enter interactive mode for follow-up questions
- `--no-context`: Skip environment context gathering
- `--dry-run`: Perform redaction and context gathering but do not call API

### Examples

```bash
# Analyze error logs
cat error.log | que

# Use Claude with verbose output
tail -f app.log | que --provider claude --verbose

# Interactive mode - ask follow-up questions
cat error.log | que -i

# Dry run to see what would be sent
cat config.yaml | que --dry-run --verbose

# Skip context gathering
cat log.txt | que --no-context

# Interactive mode with specific provider
cat server.log | que --provider claude -i
```

### CI/CD and Server Use Cases

Que is perfect for automated environments where you need AI-powered log analysis without interactive editors:

**GitHub Actions:**
```yaml
- name: Analyze deployment errors
  if: failure()
  run: |
    cat deployment.log | que --no-context > analysis.txt
    cat analysis.txt
```

**Docker/Kubernetes:**
```bash
# Analyze container logs
kubectl logs pod-name | que --no-context

# Analyze Docker logs
docker logs container-name 2>&1 | que
```

**Server Monitoring:**
```bash
# Analyze systemd service failures
journalctl -u my-service --since "1 hour ago" | que

# Analyze application errors from log files
tail -n 1000 /var/log/app/error.log | que --provider claude
```

**Automated Error Reporting:**
```bash
# Send analysis to Slack/email
cat error.log | que --no-context | mail -s "Error Analysis" admin@example.com
```

## How It Works

Que follows a linear pipeline architecture:

1. **Ingestor**: Reads from stdin (with buffer limits to prevent memory overflow)
2. **Enricher**: Gathers non-sensitive metadata from the host environment
3. **Sanitizer**: Redacts PII and secrets using gitleaks detection
4. **Advisor**: Formats the payload, selects the provider, sends the request, and renders the response

### Interactive Mode

When using the `-i` or `--interactive` flag, Que enters an interactive session after displaying the initial analysis. This allows you to:

- Ask follow-up questions about the log analysis
- Get clarification on the root cause or fix
- Request additional details or alternative solutions
- Have a conversation with the AI while maintaining full context of the original log

To exit interactive mode, type `exit`, `quit`, or `q`.

## License

MIT

