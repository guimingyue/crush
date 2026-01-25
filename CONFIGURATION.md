# Crush Configuration Guide

This guide explains how to configure Crush with a simple, streamlined approach.

## Quick Setup

The simplest way to configure Crush is to set just the essential properties in a configuration file.

### 1. Create a Configuration File

Create a file named `crush.yaml` in your home directory or working directory:

```yaml
# crush.yaml
model:
  provider: "openai"  # or "anthropic", "google", "ollama"
  name: "gpt-4o"      # or "claude-3-5-sonnet", "gemini-1.5-pro", "llama3.1:8b"
  api_key: "your-api-key-here"  # Leave empty for local models like Ollama
  base_url: ""        # Leave empty for standard providers, specify for custom endpoints
```

### 2. Environment Variables (Alternative)

You can also set these as environment variables:

```bash
export CRUSH_MODEL_PROVIDER=openai
export CRUSH_MODEL_NAME=gpt-4o
export CRUSH_API_KEY=your-api-key-here
```

### 3. Supported Providers

| Provider | Model Name Example | API Key Required |
|----------|-------------------|------------------|
| openai | gpt-4o, gpt-4o-mini | Yes |
| anthropic | claude-3-5-sonnet, claude-opus | Yes |
| google | gemini-1.5-pro, gemini-1.0-pro | Yes |
| ollama | llama3.1:8b, mistral, phi3 | No |
| groq | llama3-groq-8b-tool-use-preview | Yes |

### 4. Local Model Setup (Ollama)

For local models using Ollama:

```yaml
# crush.yaml
model:
  provider: "ollama"
  name: "llama3.1:8b"  # or any model available in Ollama
  api_key: ""           # Not needed for local models
  base_url: "http://localhost:11434"  # Default Ollama endpoint
```

Then make sure Ollama is running:

```bash
# Install and start Ollama
brew install ollama  # On macOS
ollama serve         # Start the server in another terminal
ollama pull llama3.1:8b  # Pull the model you want to use
```

### 5. Run Crush

Once configured, simply run:

```bash
./crush
```

The application will automatically use your configuration settings.

## Minimal Configuration Examples

### OpenAI GPT-4o:
```yaml
model:
  provider: "openai"
  name: "gpt-4o"
  api_key: "sk-...your-openai-api-key"
```

### Anthropic Claude:
```yaml
model:
  provider: "anthropic"
  name: "claude-3-5-sonnet"
  api_key: "your-anthropic-api-key"
```

### Google Gemini:
```yaml
model:
  provider: "google"
  name: "gemini-1.5-pro"
  api_key: "your-google-api-key"
```

### Local Ollama:
```yaml
model:
  provider: "ollama"
  name: "mistral"
  api_key: ""
  base_url: "http://localhost:11434"
```

That's it! With just these simple settings, Crush will be fully configured to work with your chosen LLM provider.