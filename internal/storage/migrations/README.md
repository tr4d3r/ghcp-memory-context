# Database Migrations

This directory contains database migration files for the GHCP Memory Context Server storage layer.

## Migration Files

Migration files follow the naming convention: `{version}_{name}.{direction}.sql`

- `version`: Sequential integer (001, 002, etc.)
- `name`: Descriptive name using snake_case
- `direction`: Either `up` or `down`

### Current Migrations

- `001_initial_schema.up.sql` - Creates the initial database schema with core tables
- `001_initial_schema.down.sql` - Removes the initial database schema

## Schema Overview

The initial schema includes:

### Core Tables

1. **sessions** - User sessions with context management
2. **context_objects** - Generic storage for all MCP context objects
3. **tasks** - Specialized task storage with denormalized fields for performance
4. **task_dependencies** - Many-to-many relationships between tasks
5. **code_references** - File and line references for tasks
6. **schema_migrations** - Migration tracking

### Key Design Decisions

- **MCP Compliance**: All context objects follow the Model Context Protocol specification
- **Hybrid Storage**: Generic context storage + specialized task storage for performance
- **Audit Trail**: Created/updated timestamps and user tracking
- **Flexible Metadata**: JSON fields for extensible key-value storage
- **Performance Optimized**: Strategic indexing for common query patterns
- **Data Integrity**: Foreign keys, constraints, and triggers

### Triggers

The schema includes triggers for:
- Automatic `updated_at` timestamp updates
- Task status change tracking (started_at, completed_at)

## Migration System

The migration system will:
1. Track applied migrations in the `schema_migrations` table
2. Execute migrations in order based on version number
3. Support both up and down migrations
4. Verify checksums to detect file modifications
5. Provide rollback capabilities

## Usage

Migrations will be executed automatically when the storage layer initializes or through explicit migration commands in the application.
