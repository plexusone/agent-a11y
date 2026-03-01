# LLM Integration

agent-a11y uses LLM-as-a-Judge to evaluate accessibility findings and reduce false positives.

## How It Works

1. axe-core identifies potential accessibility issues
2. For each finding, the LLM evaluates:
   - Is this a true accessibility barrier?
   - What is the severity?
   - What is the recommended fix?
3. Findings are filtered or annotated based on LLM evaluation

## Supported Providers

| Provider | Models | Environment Variable |
|----------|--------|---------------------|
| Anthropic | claude-sonnet-4-20250514, claude-opus-4-20250514 | `ANTHROPIC_API_KEY` |
| OpenAI | gpt-4o, gpt-4-turbo | `OPENAI_API_KEY` |
| Google | gemini-2.0-flash | `GOOGLE_API_KEY` |
| xAI | grok-2 | `XAI_API_KEY` |
| Ollama | llama3, mistral, etc. | (local) |

## Configuration

### CLI

```bash
export ANTHROPIC_API_KEY="your-key"

agent-a11y audit https://example.com \
  --llm-provider anthropic \
  --llm-model claude-sonnet-4-20250514
```

### Config File

```yaml
llm:
  enabled: true
  provider: anthropic
  model: claude-sonnet-4-20250514
```

### Go Library

```go
auditor, _ := a11y.New(
    a11y.WithLLM("anthropic", "claude-sonnet-4-20250514", apiKey),
)
```

## Evaluation Output

With LLM enabled, each finding includes:

```json
{
  "ruleId": "color-contrast",
  "description": "Elements must have sufficient color contrast",
  "llmConfirmed": true,
  "llmReasoning": "The text has a contrast ratio of 3.2:1, which fails WCAG AA requirements of 4.5:1 for normal text.",
  "llmConfidence": 0.95
}
```

## Ollama (Local)

For local LLM evaluation without API keys:

```bash
# Start Ollama
ollama serve

# Pull a model
ollama pull llama3

# Use with agent-a11y
agent-a11y audit https://example.com \
  --llm-provider ollama \
  --llm-model llama3
```

## Cost Considerations

LLM evaluation adds API costs. To manage costs:

- Use smaller models for initial scans
- Enable LLM only for final audits
- Set confidence thresholds to skip obvious issues

```yaml
llm:
  enabled: true
  provider: anthropic
  model: claude-haiku-20250514  # Cheaper model
  confidenceThreshold: 0.8      # Skip low-confidence evaluations
```
