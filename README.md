# Gaia (盖娅)

Gaia 是一个 AI 智能助手工具，类似于 Claude Code / OpenClaw，支持：
- Daily work automation (file organization, email handling)
- Code writing, modification, testing, and architecture design
- OpenAI-compatible backend model interface
- MCP protocol support
- Plugin architecture
- CLI interface (GUI planned for future)

## Quick Start

```bash
# Install dependencies
make deps

# Build
make build

# Run
./build/gaia chat

# Or run directly
make dev
```

## Configuration

Copy `configs/config.yaml.example` to `configs/config.yaml` and configure your settings:

```bash
cp configs/config.yaml.example configs/config.yaml
```

Set your API key:
```bash
export OPENAI_API_KEY=your-api-key
```

## Features

- **LLM Provider**: OpenAI-compatible interface supporting multiple backends
- **Plugin System**: Extensible plugin architecture
- **MCP Protocol**: Model Context Protocol support
- **Built-in Tools**: File operations, shell execution, code tools, email, git, web requests
- **Session Management**: Persistent conversation history with SQLite

## License

MIT
