# quill

<p align="center">
  <strong>AI-Powered Git Commit Message Generator</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#commands">Commands</a> •
  <a href="#providers">Providers</a> •
  <a href="#contributing">Contributing</a>
</p>

`quill` streamlines your git workflow by generating contextually aware, conventional commit messages using AI. It analyzes your staged changes and generates properly formatted commit messages that accurately describe your changes.

## Features

- 🤖 **AI-Powered Generation**: Intelligent commit message suggestions using state-of-the-art language models
- 🎯 **Conventional Commits**: Automatically formatted according to the conventional commits specification
- 🎨 **Interactive UI**: Beautiful terminal interface for selecting and editing commit messages
- ⚡ **Multiple Providers**: Support for Google Gemini, Anthropic Claude, OpenAI GPT-4o, and Ollama
- 🔒 **Secure**: API keys stored securely in your system's keyring
- 🚀 **Performance**: Rate limiting and retry mechanisms built-in
- ⚙️ **Configurable**: Extensive configuration options for customizing behavior

## Quick Start

```bash
# Install
go install github.com/jabafett/quill@latest

# Initialize (interactive setup)
quill init

# Generate commit messages
quill generate
```

## Commands

| Command               | Description                                 |
| --------------------- | ------------------------------------------- |
| (✅) `quill init`     | Create initial configuration                |
| (✅) `quill generate` | Generate commit message from staged changes |
| (🚧) `quill suggest`  | Suggest file groupings based on analysis    |
| (🚧) `quill history`  | Show message history                        |
| (✅) `quill config`   | Manage configuration                        |

## Providers

Supported AI providers:

- **Google Gemini**: Fast and efficient, great for everyday use
- **Anthropic Claude**: Excellent understanding of code context
- **OpenAI**: High-quality results with GPT-4
- **Ollama**: Local execution, no API costs

## Configuration

Configuration is stored in either:
- `~/.config/quill.toml`
- `~/.config/.quill.toml`

Key features:
- Configurable rate limiting (1 request/second)
- Configurable retries
- Provider-specific settings
- Multiple candidate generation

## System Requirements

### Secure Keyring Storage

quill uses your system's secure keyring to store API keys:

- **Linux**: `libsecret` (GNOME Keyring) or `kwallet` (KDE Wallet)
- **macOS**: Keychain
- **Windows**: Windows Credential Manager

## Advanced Usage

### Command Details

#### Generate Command
```bash
quill generate [flags]

Flags:
  -p, --provider string      Override default AI provider
  -c, --candidates int      Number of commit message variations (1-3)
  -t, --temperature float   Generation temperature (0.0-1.0)
```

#### Config Management
```bash
# View current configuration
quill config list

# Get specific setting
quill config get core.default_provider

# Set configuration value
quill config set providers.gemini.temperature 0.7

# Manage API keys
quill config set-key gemini YOUR_API_KEY
quill config get-key gemini
```

### Provider Configuration

Each provider can be customized in `quill.toml`:

```toml
[providers.gemini]
model = "gemini-1.5-flash-002"
max_tokens = 8192
temperature = 0.3
enable_retries = true
candidate_count = 2
```

Available settings:
- `model`: Model identifier
- `max_tokens`: Maximum response length
- `temperature`: Creativity of responses (0.0-1.0)
- `enable_retries`: Enable automatic retry on failure
- `candidate_count`: Default number of suggestions

### Rate Limiting & Retries

- Rate limiting: 1 request per second
- Retries: Up to 3 attempts with exponential backoff
- Configurable per provider via `enable_retries`

### Environment Variables

While API keys are preferably stored in the system keyring, they can be provided via environment variables:

```bash
export GEMINI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
```

### Interactive UI Controls

When selecting commit messages:
- `↑/↓` or `j/k`: Navigate options
- `enter`: Select message
- `e`: Edit selected message
- `q`: Quit without selecting

### Git Integration

quill works with staged changes:
```bash
# Stage specific files
git add file1.go file2.go

# Stage all changes
git add .

# Generate commit message
quill generate

```

### Upcoming Features

- `quill suggest`: Analyze changes and suggest logical commit groupings
- `quill history`: Track and reuse previous commit messages
- Pre-commit hook integration
- IDE extensions

## Contributing

See [Contributing Guidelines](docs/CONTRIBUTING.md) and [Roadmap](docs/ROADMAP.md).

## License

MIT License - see [LICENSE](LICENSE) for details

## Credits

This project was inspired by:

- [Conventional Commits](https://www.conventionalcommits.org)
- [commitgpt](https://github.com/RomanHotsiy/commitgpt)
- [A Blog by Harper](https://harper.blog/2024/03/11/use-an-llm-to-automagically-generate-meaningful-git-commit-messages/)
