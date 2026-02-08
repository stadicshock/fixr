# fixr

**Fix grammar anywhere you write — powered by AI models you already pay for.**

You already have access to powerful AI models through paid API keys (Anthropic, OpenAI), your company's cloud accounts (AWS Bedrock), or locally via Ollama. fixr makes it dead simple to use those models for everyday grammar correction — right from your clipboard, in one command.

## Why fixr?

AI models like Claude and GPT-4o are excellent at fixing grammar. But using them means opening a browser, pasting into a chat window, waiting, and copying the result back. Every time.

fixr removes that friction. One command, and your clipboard text is fixed.

It also keeps things secure. Your text goes directly to the AI provider you configure — your own API key, your own AWS account. No middleman, no free-tier services storing your data, no browser extensions reading everything you type.

- **Use AI you already pay for**: Bring your own API keys — Anthropic, OpenAI, AWS Bedrock, or local Ollama
- **One command**: Copy text, run `fixr fix -c`, paste the corrected version
- **Context-aware**: Optionally captures a screenshot of your active window so the AI understands conversation tone and setting
- **Single binary**: One `go build`, one executable. No runtime, no dependencies
- **Secure by design**: Text goes directly to your configured provider. No data collection, no third-party access

## Quick Start

```bash
# Build from source
git clone https://github.com/stadicshock/fixr && cd fixr
go build -o fixr .

# Create config
./fixr config init
# Edit ~/.fixr/config.yaml with your API key

# Fix clipboard text
./fixr fix -c
```

## How It Works

1. **Copy** text you want to fix (Cmd+C)
2. **Run** `fixr fix -c`
3. **Paste** the corrected text (Cmd+V)

fixr automatically captures a screenshot of your active window for context — the AI sees the conversation thread and adjusts tone accordingly.

## Setting Up a Global Hotkey

To trigger fixr from anywhere with a keyboard shortcut:

1. Open **Automator** > New > **Quick Action**
2. Set "Workflow receives" to **no input**
3. Add **Run Shell Script** action
4. Enter the full path to fixr:
   ```
   /path/to/fixr fix -c
   ```
5. Save as "Fix Grammar"
6. Go to **System Settings > Keyboard > Keyboard Shortcuts > Services**
7. Find "Fix Grammar" and assign **Ctrl+Shift+G** (or your preferred shortcut)

Now from anywhere: copy text, press the hotkey, paste the corrected version.

## Configuration

Config lives at `~/.fixr/config.yaml`. Create it with:

```bash
fixr config init
```

### Example config

```yaml
default_provider: bedrock

providers:
  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-sonnet-4-20250514

  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4o

  bedrock:
    region: us-east-1
    model_id: us.anthropic.claude-3-5-haiku-20241022-v1:0
    profile: my-aws-profile      # AWS profile from ~/.aws/config

  ollama:
    host: http://localhost:11434
    model: llama3

preferences:
  tone: professional              # professional | casual | formal
  language: en-US
  auto_screenshot: true           # Capture active window for context
  auto_paste: false               # Auto-paste corrected text
```

### API Keys

Set your API keys as environment variables (recommended) or directly in the config:

```bash
# Add to your ~/.zshrc or ~/.bashrc
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
```

### Providers

| Provider | Vision | Local | Setup |
|----------|--------|-------|-------|
| **Anthropic** | Yes | No | Set `ANTHROPIC_API_KEY` |
| **OpenAI** | Yes | No | Set `OPENAI_API_KEY` |
| **AWS Bedrock** | Depends on model | No | Configure `~/.aws/credentials` + `profile` |
| **Ollama** | Model-dependent | Yes | Install and run Ollama |

## Usage

```bash
# Fix clipboard text, copy result back
fixr fix -c

# Fix clipboard text, print to stdout
fixr fix

# Fix a file
fixr fix -f draft.txt

# Use a specific provider
fixr fix -p openai -c

# Show current config (keys masked)
fixr config show

# Show config file path
fixr config path
```

## Building

```bash
make build     # Build binary
make test      # Run tests
make install   # Install to $GOPATH/bin
make clean     # Remove build artifacts
```

## Project Structure

```
fixr/
├── main.go                 # Entry point
├── cmd/                    # CLI commands (cobra)
│   ├── root.go
│   ├── fix.go              # fixr fix
│   └── config.go           # fixr config
├── internal/
│   ├── config/             # YAML config management
│   ├── engine/             # Core orchestrator
│   ├── capture/            # Screenshot capture (macOS screencapture)
│   ├── clipboard/          # Clipboard read/write (pbcopy/pbpaste)
│   ├── prompt/             # AI prompt templates
│   └── provider/           # AI provider implementations
│       ├── provider.go     # Provider interface
│       ├── anthropic.go    # Anthropic Claude
│       ├── openai.go       # OpenAI GPT-4o
│       ├── bedrock.go      # AWS Bedrock
│       └── ollama.go       # Ollama (local)
├── config.example.yaml
├── Makefile
└── LICENSE
```

## License

MIT
