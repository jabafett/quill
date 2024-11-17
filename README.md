# Quill

Quill streamlines the git commit workflow by generating contextually aware, conventional commit messages using AI. It analyzes your staged changes and generates properly formatted commit messages that accurately describe your changes. Built with flexibility in mind, Quill supports multiple AI providers and offers both CLI and interactive interfaces to suit different workflows.

## Features

- 🤖 AI-powered commit message generation
- 🎨 Interactive UI with message selection
- ⚙️ Configurable AI providers and settings
- 📝 Conventional commit format support
- 🔄 Progress indicators and error handling
- 🛠️ Easy configuration management

## Quick Start

```bash
go install github.com/jabafett/quill@latest

quill init

# Be inside a git repository
quill generate
```

## Commands

| Command                  | Description                                 |
| ------------------------ | ------------------------------------------- |
| (✅) `quill init`        | Create initial configuration                |
| (✅) `quill generate`    | Generate commit message from staged changes |
| (🚧) `quill suggest`     | Suggest file groupings based on analysis    |
| (🚧) `quill history`     | Show message history                        |
| (✅) `quill config`      | Manage configuration                        |

## Configuration

Quill uses a TOML configuration file located at `~/.quill/config.toml`. Key settings include:

```toml
[core]
default_provider = "gemini"
cache_ttl = "24h"
[providers.gemini]
model = "gemini-1.5-flash"
temperature = 0.3
```

## Currently Supported Providers

- ✅ Google Gemini
- 🚧 OpenAI (planned)
- 🚧 Anthropic (planned)
- 🚧 Ollama (planned)

### Key Components

#### AI Provider System

- Provider interface with pluggable implementations
- Configurable model parameters and settings
- Built-in retry mechanisms and rate limiting
- Response validation and error recovery
- Context-aware prompt management

#### Git Integration

- Efficient staged changes detection
- Intelligent diff analysis and parsing
- Breaking change detection
- Scope inference from file paths
- Multi-file context awareness

#### Configuration Management

- Hierarchical configuration system
- Environment variable integration
- Secure credentials handling
- Provider-specific settings
- Runtime configuration updates

#### Performance Features

- Optimized diff processing
- Smart caching with TTL
- Concurrent request handling
- Memory-efficient operations
- Request deduplication

## Dependencies

### System Requirements

#### Secure Keyring Storage
Quill uses the system keyring to securely store API keys. This requires:

- **Linux**: `libsecret` (GNOME Keyring) or `kwallet` (KDE Wallet)
- **macOS**: Keychain
- **Windows**: Windows Credential Manager

If the system keyring is not available, Quill will fall back to using environment variables.

### Core

- [go-git/go-git](https://github.com/go-git/go-git)
- [spf13/cobra](https://github.com/spf13/cobra)
- [spf13/viper](https://github.com/spf13/viper)
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)

## Credits

This project is inspired by:

- [Conventional Commits](https://www.conventionalcommits.org)
- [commitgpt](https://github.com/RomanHotsiy/commitgpt)
- [GitPilotAI](https://github.com/ksred/GitPilotAI)

## Contributing

Contributions are welcome! See our [ROADMAP.md](docs/ROADMAP.md) for planned features.
