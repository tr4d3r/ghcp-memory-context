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

## Installation & Setup

### Option 1: Download Pre-built Binary (Recommended)

**Automatic Installation (Linux/macOS)**
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/tr4d3r/ghcp-memory-context/main/scripts/install.sh | bash

# Or install specific version
curl -fsSL https://raw.githubusercontent.com/tr4d3r/ghcp-memory-context/main/scripts/install.sh | bash -s v1.0.0

# Run the server
ghcp-memory-context
```

**Manual Download**
```bash
# Linux x64
curl -L https://github.com/tr4d3r/ghcp-memory-context/releases/latest/download/ghcp-memory-context-linux-amd64.tar.gz | tar xz
chmod +x ghcp-memory-context
./ghcp-memory-context

# Linux ARM64
curl -L https://github.com/tr4d3r/ghcp-memory-context/releases/latest/download/ghcp-memory-context-linux-arm64.tar.gz | tar xz

# macOS x64 (Intel)
curl -L https://github.com/tr4d3r/ghcp-memory-context/releases/latest/download/ghcp-memory-context-darwin-amd64.tar.gz | tar xz

# macOS ARM64 (Apple Silicon)
curl -L https://github.com/tr4d3r/ghcp-memory-context/releases/latest/download/ghcp-memory-context-darwin-arm64.tar.gz | tar xz

# Windows x64 (PowerShell)
Invoke-WebRequest -Uri "https://github.com/tr4d3r/ghcp-memory-context/releases/latest/download/ghcp-memory-context-windows-amd64.zip" -OutFile "ghcp-memory-context.zip"
Expand-Archive -Path "ghcp-memory-context.zip" -DestinationPath "."
.\ghcp-memory-context.exe
```

**Package Managers**
```bash
# Homebrew (macOS/Linux)
brew install tr4d3r/tap/ghcp-memory-context

# Chocolatey (Windows)
choco install ghcp-memory-context

# Scoop (Windows)
scoop bucket add tr4d3r https://github.com/tr4d3r/scoop-bucket
scoop install ghcp-memory-context

# APT (Debian/Ubuntu)
curl -fsSL https://apt.tr4d3r.dev/gpg | sudo apt-key add -
echo "deb https://apt.tr4d3r.dev stable main" | sudo tee /etc/apt/sources.list.d/tr4d3r.list
sudo apt update && sudo apt install ghcp-memory-context

# YUM/DNF (RHEL/CentOS/Fedora)
sudo rpm --import https://yum.tr4d3r.dev/gpg
sudo yum-config-manager --add-repo https://yum.tr4d3r.dev/tr4d3r.repo
sudo yum install ghcp-memory-context
```

**System Service Installation**
```bash
# Install as systemd service (Linux)
sudo ./ghcp-memory-context install --service
sudo systemctl enable ghcp-memory-context
sudo systemctl start ghcp-memory-context

# Install as service (Windows)
.\ghcp-memory-context.exe install --service
sc start GHCPMemoryContext

# Install as service (macOS)
sudo ./ghcp-memory-context install --service
sudo launchctl load /Library/LaunchDaemons/com.tr4d3r.ghcp-memory-context.plist
```

### Option 2: Docker (Quick Start)

**Pre-built Image (Recommended)**
```bash
# Pull and run the Docker image
docker run -d \
  --name ghcp-memory-context \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  tr4d3r/ghcp-memory-context:latest

# Verify it's running
curl http://localhost:8080/health
```

**Build from Source**
```bash
# Clone repository
git clone https://github.com/tr4d3r/ghcp-memory-context.git
cd ghcp-memory-context

# Build Docker image
docker build -t ghcp-memory-context:local .

# Run container
docker run -d \
  --name ghcp-memory-context \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ghcp-memory-context:local
```

**Docker Compose (Recommended for Development)**
```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Option 3: Build from Source

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

## VS Code & GitHub Copilot Premium Setup

### Prerequisites
- **VS Code** with **GitHub Copilot** extension installed
- **GitHub Copilot Premium** subscription (required for MCP support)
- Memory Context Server running on `localhost:8080`

### Step 1: Install VS Code Extensions

```bash
# Install required extensions
code --install-extension GitHub.copilot
code --install-extension GitHub.copilot-chat
```

### Step 2: Configure MCP Integration

Create or update your VS Code settings to include the memory context server:

**Method A: VS Code Settings JSON** (`.vscode/settings.json`)
```json
{
  "github.copilot.enable": {
    "*": true
  },
  "github.copilot.advanced": {
    "mcp": {
      "servers": {
        "memory-context": {
          "command": "http",
          "args": ["http://localhost:8080/mcp"],
          "env": {}
        }
      }
    }
  }
}
```

**Method B: MCP Configuration File** (`.vscode/mcp.json`)
```json
{
  "mcpServers": {
    "memory-context": {
      "command": "http",
      "args": ["http://localhost:8080/mcp"],
      "description": "Persistent memory context for project development"
    }
  }
}
```

### Step 3: Start the Memory Server

```bash
# Start the memory context server
./ghcp-memory-context

# Or with Docker
docker run -d -p 8080:8080 -v $(pwd)/data:/app/data tr4d3r/ghcp-memory-context:latest

# Verify server is running
curl http://localhost:8080/health
```

### Step 4: Verify Integration

1. **Open VS Code** in your project directory
2. **Open Copilot Chat** (Ctrl/Cmd + Shift + P → "GitHub Copilot: Open Chat")
3. **Test memory commands**:
   ```
   Remember that we use conventional commits in this project
   What coding standards should I follow for this project?
   Search for any authentication patterns we've used
   ```

### Step 5: Using Memory Context in Development

Once configured, Copilot will automatically:
- **Remember** your coding patterns and decisions
- **Recall** project-specific context across sessions
- **Search** through stored knowledge when providing suggestions

**Manual memory operations via Copilot Chat:**
```
@memory remember: "Use TypeScript strict mode for all new files"
@memory recall: project_standards
@memory search: "authentication patterns"
```

### Troubleshooting VS Code Integration

**Issue: Copilot not connecting to memory server**
```bash
# Check server status
curl http://localhost:8080/health

# Check MCP resources
curl http://localhost:8080/mcp/resources

# Restart VS Code and reload window
```

**Issue: Memory commands not working**
1. Verify GitHub Copilot Premium subscription
2. Check VS Code extension versions
3. Ensure MCP configuration is properly formatted
4. Restart the memory context server

## Alternative Editor Setup

### Cursor IDE
```json
// .cursor/mcp.json
{
  "mcpServers": {
    "memory-context": {
      "command": "http",
      "args": ["http://localhost:8080/mcp"]
    }
  }
}
```

### Neovim with Copilot
```lua
-- init.lua or copilot.lua
require('copilot').setup({
  mcp = {
    servers = {
      ["memory-context"] = {
        command = "http",
        args = {"http://localhost:8080/mcp"}
      }
    }
  }
})
```

### JetBrains IDEs (IntelliJ, PyCharm, etc.)
```xml
<!-- .idea/mcp.xml -->
<mcpConfiguration>
  <server name="memory-context">
    <command>http</command>
    <args>http://localhost:8080/mcp</args>
  </server>
</mcpConfiguration>
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

## Docker Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DATA_DIR` | `/app/data` | Data storage directory |

### Volume Mounts

| Host Path | Container Path | Description |
|-----------|----------------|-------------|
| `./data` | `/app/data` | Persistent memory storage |

### Advanced Docker Usage

**Custom Port and Data Directory**
```bash
docker run -d \
  --name ghcp-memory-context \
  -p 3000:3000 \
  -e PORT=3000 \
  -e DATA_DIR=/app/custom-data \
  -v /host/path/data:/app/custom-data \
  tr4d3r/ghcp-memory-context:latest
```

**Production Deployment**
```bash
# Create named volume for data persistence
docker volume create ghcp-data

# Run with named volume and restart policy
docker run -d \
  --name ghcp-memory-context \
  --restart unless-stopped \
  -p 8080:8080 \
  -v ghcp-data:/app/data \
  --memory=512m \
  --cpus=0.5 \
  tr4d3r/ghcp-memory-context:latest
```

**Health Monitoring**
```bash
# Check container health
docker ps --filter name=ghcp-memory-context

# View health check logs
docker inspect ghcp-memory-context --format='{{json .State.Health}}'

# Manual health check
docker exec ghcp-memory-context wget --no-verbose --tries=1 --spider http://localhost:8080/health
```

### Docker Compose for Production

Create `docker-compose.prod.yml`:
```yaml
version: '3.8'

services:
  ghcp-memory-context:
    image: tr4d3r/ghcp-memory-context:latest
    container_name: ghcp-memory-context
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - DATA_DIR=/app/data
    volumes:
      - ghcp_data:/app/data
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '0.50'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

volumes:
  ghcp_data:
    driver: local
```

## Development

```bash
# Run tests
go test ./...

# Run with verbose logging
go run cmd/server/main.go

# Build for production
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Build Docker image for development
docker build -t ghcp-memory-context:dev .

# Test Docker build
docker run --rm -p 8080:8080 ghcp-memory-context:dev
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
