# GHCP Memory Context Server

A Model Context Protocol (MCP) compliant memory server for GitHub Copilot Premium integration, providing persistent context storage and atomic fact memory for AI-assisted development.

## Overview

This project implements a **memory/context server** that provides persistent, structured memory for AI applications. It acts as the "memory infrastructure" for GitHub Copilot Premium, significantly enhancing AI assistant capabilities through persistent context across development sessions.

**Core Functionality**: Store and retrieve atomic facts, entity relationships, and contextual information that AI assistants can access via standardized MCP protocol.

## Key Value Proposition

**Enhanced GitHub Copilot Premium (GHCP) in VS Code:**
- Persistent project context across development sessions  
- Memory of coding standards, architectural decisions, and patterns
- Cross-session continuity enabling AI-assisted development workflows
- Project-specific context enriching Copilot conversations

## Requirements

- **Go 1.23.10 or later** (required for security patches)
- No external database dependencies (uses JSON file storage)

## Architecture

- **Memory Context Server**: Persistent storage for atomic facts and entity relationships
- **Entity-Relationship Model**: Entities with observations (facts) and inter-entity relations
- **File-Based Storage**: JSON files with in-memory caching for performance
- **MCP Protocol Compliance**: Full integration with AI assistants via standardized interface
- **RESTful API**: Direct HTTP access alongside MCP protocol support

## Core Concepts

### Entity-Relationship Memory Model
```
Entity: project_standards (type: guideline)
├── Observation: "use conventional commits"
├── Observation: "format: type(scope): description"  
└── Observation: "breaking changes go in footer"

Relation: project → follows → project_standards
```

### Atomic Facts Storage
- **Entities**: Named containers for related context (e.g., "coding_standards", "api_patterns")
- **Observations**: Individual, atomic facts about entities (e.g., "use REST endpoints")
- **Relations**: Connections between entities (e.g., "project follows coding_standards")

## Features

- ✅ **Memory Operations**: "remember X", "recall Y", "search Z" functionality
- ✅ **Entity Management**: CRUD operations for contextual entities
- ✅ **Atomic Facts**: Granular observation storage and retrieval
- ✅ **Cross-Entity Search**: Find facts across all stored context
- ✅ **MCP Protocol**: Full compliance for AI assistant integration
- ✅ **RESTful API**: HTTP endpoints for direct integration
- ✅ **File-Based Storage**: No database setup required
- ✅ **Concurrent Access**: Thread-safe operations with file locking
- ✅ **JSON-Direct Serving**: Optimized for MCP's JSON-RPC protocol

## Quick Start

```bash
# Clone the repository
git clone https://github.com/tr4d3r/ghcp-memory-context.git
cd ghcp-memory-context

# Build and run the server
go build -o bin/server cmd/server/main.go
./bin/server

# Or run directly with custom data directory
go run cmd/server/main.go /path/to/data

# Server will start on http://localhost:8080
```

## API Usage Examples

### Core Memory Operations

```bash
# Remember a fact
curl -X POST http://localhost:8080/memory/remember \
  -H "Content-Type: application/json" \
  -d '{
    "entityName": "project_standards",
    "entityType": "guideline",
    "observation": "use conventional commits"
  }'

# Recall facts
curl http://localhost:8080/memory/recall?entity=project_standards

# Search memory
curl http://localhost:8080/memory/search?q=commit
```

### Entity Management

```bash
# List all entities
curl http://localhost:8080/entities

# Get specific entity
curl http://localhost:8080/entities/project_standards

# Create entity with observations
curl -X POST http://localhost:8080/entities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api_patterns",
    "entityType": "pattern",
    "observations": ["use REST endpoints", "implement pagination"]
  }'
```

### MCP Protocol Integration

```bash
# List MCP resources
curl http://localhost:8080/mcp/resources

# Use MCP tools
curl -X POST http://localhost:8080/mcp/tools/remember_fact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "remember_fact",
    "arguments": {
      "entityName": "coding_style",
      "observation": "prefer explicit error handling"
    }
  }'
```

## MCP Integration for Copilot

Configure in `.vscode/mcp.json` or similar MCP client:

```json
{
  "mcpServers": {
    "memory-context": {
      "command": "http",
      "args": ["http://localhost:8080/mcp"]
    }
  }
}
```

## Project Structure

```
├── cmd/
│   └── server/              # Main server application
├── internal/
│   ├── api/                 # RESTful API handlers
│   │   ├── router.go        # Main routing and middleware
│   │   ├── entities.go      # Entity CRUD endpoints
│   │   ├── memory.go        # Memory operations (remember/recall/search)
│   │   ├── relations.go     # Entity relationship management
│   │   ├── mcp.go           # MCP protocol endpoints
│   │   └── utils.go         # Common utilities and response handling
│   ├── models/              # Data models
│   │   └── entity.go        # Entity, Observation, Relation models
│   └── storage/
│       ├── interface.go     # Storage abstraction interfaces
│       └── filestore/       # JSON file-based storage implementation
├── pkg/
│   └── types/               # Context object types
└── data/                    # Default data directory (created at runtime)
    ├── entities/            # Individual entity JSON files
    └── relations/           # Relations JSON file
```

## Storage Format

### Entity Files (`data/entities/{name}.json`)
```json
{
  "name": "project_standards",
  "entityType": "guideline",
  "createdAt": "2025-08-20T10:30:00Z",
  "lastModified": "2025-08-20T15:45:00Z",
  "observations": [
    {
      "id": "obs_001",
      "text": "use conventional commits",
      "createdAt": "2025-08-20T10:30:00Z",
      "source": "user_input"
    }
  ]
}
```

### Relations File (`data/relations/relations.json`)
```json
{
  "relations": [
    {
      "id": "rel_001",
      "from": "current_project",
      "to": "project_standards",
      "relationType": "follows",
      "createdAt": "2025-08-20T10:30:00Z"
    }
  ]
}
```

## Configuration

### Environment Variables
- `PORT`: Server port (default: 8080)
- `DATA_DIR`: Data storage directory (default: ./data)

### Command Line
```bash
# Custom data directory
go run cmd/server/main.go /custom/data/path

# With environment variables
PORT=3000 DATA_DIR=/var/lib/memory-context go run cmd/server/main.go
```

## API Reference

### Memory Operations
- `POST /memory/remember` - Store atomic facts
- `GET /memory/recall` - Retrieve stored context
- `GET /memory/search` - Search across all memory

### Entity Management  
- `GET /entities` - List entities
- `POST /entities` - Create entities
- `GET /entities/{name}` - Get specific entity
- `PUT /entities/{name}` - Update entity
- `DELETE /entities/{name}` - Remove entity

### Relations
- `GET /relations` - List entity relationships
- `POST /relations` - Create relationships
- `PUT /relations/{id}` - Update relationships
- `DELETE /relations/{id}` - Remove relationships

### MCP Protocol
- `GET /mcp/resources` - List available resources
- `GET /mcp/resources/{uri}` - Get resource content
- `POST /mcp/tools/*` - Execute MCP tools

## Development

```bash
# Run tests
go test ./...

# Run with verbose logging
go run cmd/server/main.go

# Build for production
go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

## Security

This project maintains up-to-date dependencies to address security vulnerabilities:

- **Go 1.23.10+**: Addresses [GO-2025-3750](https://pkg.go.dev/vuln/GO-2025-3750) - syscall vulnerability on Windows
- **golang.org/x/net v0.38.0+**: Addresses [GO-2025-3595](https://pkg.go.dev/vuln/GO-2025-3595) - XSS vulnerability

Run `govulncheck ./...` to check for any new vulnerabilities.

## Performance

- **Sub-100ms response times** for local operations
- **In-memory caching** for frequently accessed entities
- **Concurrent file access** with proper locking
- **Optimized for small-to-medium datasets** (hundreds to thousands of facts)

## What This IS

- **Memory Infrastructure**: Persistent storage for context objects
- **MCP Compliant**: Follows established protocol patterns  
- **Integration Backend**: Serves multiple client types
- **Local-First**: Optimized for individual developer workflows

## What This IS NOT

- **Task Manager**: Doesn't create or manage task logic
- **AI Agent**: Doesn't decompose or prioritize tasks
- **Planning Tool**: Doesn't provide strategic planning capabilities
- **UI Application**: Backend service only

## License

MIT License - see [LICENSE](LICENSE) for details.
