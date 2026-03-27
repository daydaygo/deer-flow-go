# DeerFlow Go

A Go implementation of the DeerFlow backend - a full-stack "super agent harness" built with:
- **go-zero** - Web framework
- **Eino** - Agent framework  
- **Viper** - Configuration management

This project is a Go rewrite of the [deer-flow](https://github.com/bytedance/deer-flow) Python backend, providing:
- Full API compatibility with the existing frontend
- LangGraph-compatible API endpoints
- High-performance streaming responses

## Prerequisites

- **Go 1.22+** - Install from [golang.org](https://golang.org/dl/)
- **Make** - Usually pre-installed on Linux/macOS

## Project Structure

```
deer-flow-go/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handler/         # HTTP handlers
│   ├── logic/           # Business logic
│   ├── agent/           # Agent orchestration
│   ├── llm/             # LLM client abstractions
│   │   └── providers/   # Provider implementations
│   ├── store/           # Data storage
│   └── model/           # Data models
├── pkg/
│   └── utils/           # Shared utilities
├── go.mod
├── Makefile
└── config.example.yaml
```

## Quick Start

1. Copy the example configuration:
   ```bash
   cp config.example.yaml config.yaml
   ```

2. Set your API key:
   ```bash
   export OPENAI_API_KEY=your-api-key
   ```

3. Run the server:
   ```bash
   make run
   ```

## Development

```bash
make build    # Build the server binary
make test     # Run tests
make clean    # Clean build artifacts
```

## Configuration

See `config.example.yaml` for configuration options.

## License

MIT