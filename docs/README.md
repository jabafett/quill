# ![quill](https://github.com/jabafett/quill/blob/main/docs/logo/quill-full-logo-50.png?raw=true, "quill") quill

## **Release 0.1.0 🎉**

Currently in development, only: `init`, `generate`, `config` commands are available.

`quill` streamlines your git workflow by generating contextually aware, conventional commit messages using AI. It analyzes your staged changes and generates properly formatted commit messages that accurately describe your changes.

[Features](#features) &middot;
[Quick Start](#quick-start) &middot;
[Commands](#commands) &middot;
[Contributing](#contributing)

## Features

- 🤖 **AI-Powered Generation**: Intelligent commit message suggestions using state-of-the-art language models
- 🎯 **Conventional Commits**: Automatically formatted according to the conventional commits specification
- 🎨 **Interactive UI**: Beautiful terminal interface for selecting and editing commit messages, including a fully-featured suggest UI for commit groupings
- ⚡ **Multiple Providers**: Support for OpenAI, Gemini, Anthropic, and Ollama
- 🔒 **Secure**: API keys stored securely in your system's keyring
- 🚀 **Performance**: Rate limiting and retry mechanisms built-in
- ⚙️ **Configurable**: Extensive configuration options for customizing behavior
- 🧩 **Commit Grouping Suggestions**: Use `quill suggest` to analyze changes and get AI-powered suggestions for logical commit groups, with auto-staging and commit support

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
| (✅) `quill suggest`  | Suggest logical commit groupings            |
| (🚧) `quill index`    | Index repository context                    |
| (🚧) `quill history`  | Show message history                        |
| (✅) `quill config`   | Manage configuration                        |

## Release Notes

### Version 0.1.0 🎉 (Initial Release)

This is the initial release of Quill, focusing on core functionality and provider integration. The following features are now available:

#### Core Features
- 🛠️ Basic Commands: `init`, `generate`, and `config` commands are fully implemented
- 🤖 Multi-Provider Support:
  - Google Gemini
  - Anthropic Claude
  - OpenAI
  - Ollama
- 🔐 Secure Configuration:
  - API keys stored in system keyring
  - TOML-based configuration
  - Provider-specific settings
- 🎨 Interactive UI:
  - Beautiful terminal interface
  - Message selection
  - Message editing

#### Technical Improvements
- ⚡ Performance Features:
  - Rate limiting (1 request/second)
  - Retry mechanism with backoff
  - File-level context caching
  - Memory-efficient processing
- 🔍 Context Analysis:
  - File-level context extraction
  - Multi-language support
  - Code symbol extraction
  - Import/dependency mapping
  - Cross-reference tracking

#### Known Limitations
- Only basic commands (`init`, `generate`, `config`) are available
- Advanced features like indexing and smart suggestions are still in development
- Some context features (historical analysis, contributor tracking) are postponed

#### Installation
```bash
go install github.com/jabafett/quill@v0.1.0

# Initialize (interactive setup)
quill init
```

## Configuration

Configuration is stored in either:

- `~/.config/quill.toml`
- `~/.config/.quill.toml`

Key features:

- Provider-specific settings
- Multiple candidate generation
- Editing generated messages
- Configurable rate limiting (1 request/second)
- Configurable retries

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


### Environment Variables

While API keys are preferably stored in the system keyring, they can be provided via environment variables:

```bash
export GEMINI_API_KEY="your-key"
export ANTHROPIC_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
```

### Interactive UI Controls

#### Commit Message Selection UI
- `↑/↓` or `j/k`: Navigate options
- `enter`: Select message and create commit
- `e`: Edit message before commit
- `q`: Quit without committing

#### Suggest Command UI
- `↑/↓` or `j/k`: Navigate suggestions
- `enter`: Select a suggestion group
- `e`: Edit the suggested commit message
- `s`: Mark a group for staging (auto-stage & commit)
- `u`: Unmark a group for staging
- `q`: Quit suggest UI
- Side panel: Shows details and files for the selected group
- Card-based layout and dynamic resizing for enhanced usability

### Upcoming Features

- `quill history`: Track and reuse previous commit messages
- Pre-commit hook integration
- IDE extensions

## Contributing

See [Roadmap](ROADMAP.md).

## License

Copyright [2024] [jabafett]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file or repository except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Credits

This project was inspired by:

- [Conventional Commits](https://www.conventionalcommits.org)
- [commitgpt](https://github.com/RomanHotsiy/commitgpt)
- [A Blog by Harper](https://harper.blog/2024/03/11/use-an-llm-to-automagically-generate-meaningful-git-commit-messages/)
