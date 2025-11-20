# Que

> The pipe-able DevOps assistant.

Que is a CLI utility designed to act as a filter in a Unix pipeline. It ingests stdin (logs, error tracebacks, config files), sanitizes the data for security, enriches it with local system context, and queries an LLM (ChatGPT or Claude) to determine the root cause and suggest a fix.

## Installation

### From Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/jenian/que.git
cd que

# Build with Make (automatically detects version from git tags)
make build

# Or install directly
make install
```

### Using Go Install

```bash
go install github.com/jenian/que/cmd/que@latest
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

