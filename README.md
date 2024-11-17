# Quill

`quill` is a CLI tool that leverages `git diff` and AI to intelligently generate commit messages or groups of commits. It supports multiple AI providers (Anthropic, OpenAI, Google, Ollama) with seamless configuration-driven functionality. 'quill' provides a simple interface for generating commit messages and file groupings. `quill` can generate multiple variants of commit messages and allow the user to choose their perferred message. Also includes the ability to generate multiple commit messages and file groupings for non-staged diffs. For now, only Gemini is supported. 

## Commands
| Command | Description |
|---------|-------------|
| `quill init` | First-time setup |
| `quill generate ` | Generate commit messages from staged/unstaged changes |
| `quill suggest` | Suggest file groupings based on analysis |
| `quill history` | Show message history |
| `quill config` | Update settings |

## Project Overview

### Core Features
- **Multi-Provider Support**: Flexible integration with various AI providers
- **Smart Diff Analysis**: Intelligent parsing and grouping of changes
- **Conventional Commits**: Enforced commit message standards
- **Caching**: Efficient message storage and retrieval
- **Configuration Management**: Easy setup and customization

### Key Components

#### AI Provider System
- Modular provider interface
- Provider-specific implementations
- Configurable models and parameters
- Robust error handling
- Rate limiting and retries

#### Git Integration
- Staged/unstaged change detection
- Smart diff parsing
- File grouping algorithms
- Change context generation

#### Configuration
- YAML/TOML based settings
- Environment variable support
- Provider-specific configs
- Secure credential management

#### Cache System
- File-based caching
- LRU eviction
- Invalidation rules
- History tracking

## Dependencies

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
