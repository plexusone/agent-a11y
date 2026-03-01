# Installation

## Go Install

The easiest way to install agent-a11y:

```bash
go install github.com/plexusone/agent-a11y/cmd/agent-a11y@latest
```

## From Source

```bash
git clone https://github.com/plexusone/agent-a11y.git
cd agent-a11y
go build -o agent-a11y ./cmd/agent-a11y
```

## Prerequisites

### Go Version

agent-a11y requires Go 1.25 or later.

### Browser

For browser-based testing, you need Chrome or Chromium installed. agent-a11y uses the Vibium browser automation library which manages browser instances.

The browser is discovered in this order:

1. `VIBIUM_CLICKER_PATH` environment variable
2. System PATH
3. Default installation locations

### LLM API Keys (Optional)

For LLM-enhanced evaluation, set the appropriate API key:

```bash
# Anthropic
export ANTHROPIC_API_KEY="your-key"

# OpenAI
export OPENAI_API_KEY="your-key"

# Google
export GOOGLE_API_KEY="your-key"

# xAI
export XAI_API_KEY="your-key"
```

## Verify Installation

```bash
agent-a11y version
```

Expected output:

```
agent-a11y v0.1.0 (commit: abc1234, built: 2026-02-28)
```

## Go Library

To use agent-a11y as a Go library:

```bash
go get github.com/plexusone/agent-a11y@latest
```

Then import:

```go
import a11y "github.com/plexusone/agent-a11y"
```
