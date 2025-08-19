# GHCP Memory Context Server

A Model Context Protocol (MCP) compliant memory server for GitHub Copilot Premium integration, providing persistent context storage for developer workflows.

## Overview

This project implements a memory/context server that bridges **Taskmaster AI** (strategic planning) with **GitHub Copilot Premium** (execution), enabling persistent, cross-session context for complex development tasks.

### Architecture

- **MCP Memory Server**: Persistent context storage with RESTful API
- **Context Objects**: Task, Code, and Chat context types
- **Local Storage**: SQLite for MVP, extensible to shared deployments
- **VS Code Integration**: Extension for context capture and injection
- **Taskmaster AI**: Strategic task planning and decomposition

## Features

- ✅ MCP-compliant context object storage
- ✅ RESTful API for CRUD operations
- ✅ Session/user/project scoping
- ✅ Semantic versioning support
- ✅ Local SQLite storage (MVP)
- 🔄 VS Code extension integration
- 🔄 Taskmaster AI MCP integration
- 🔄 GitHub Projects/Issues sync

## Quick Start

```bash
# Clone the repository
git clone https://github.com/tr4d3r/ghcp-memory-context.git
cd ghcp-memory-context

# Build and run the server
go build -o bin/server cmd/server/main.go
./bin/server

# Or run directly
go run cmd/server/main.go
```

## Project Structure

```
├── cmd/
│   └── server/           # Main server application
├── internal/
│   ├── api/             # REST API handlers
│   ├── models/          # Data models and schemas
│   └── storage/         # Database and persistence layer
├── pkg/
│   ├── mcp/             # MCP protocol implementation
│   └── types/           # Public types and interfaces
├── docs/                # Documentation
└── examples/            # Usage examples
```

## Development

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.