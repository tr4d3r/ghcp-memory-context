-- Initial schema for GHCP Memory Context Server
-- This migration creates the core tables for tasks, contexts, and sessions

-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Sessions table - represents user sessions with context
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    project_id TEXT,
    name TEXT,
    description TEXT,
    metadata TEXT, -- JSON string for key-value pairs
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    last_accessed_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    expires_at DATETIME
);

-- Create indexes for sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_project_id ON sessions(project_id);
CREATE INDEX idx_sessions_active ON sessions(active);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_last_accessed ON sessions(last_accessed_at);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Context objects table - generic storage for all context types (tasks, code, chat)
CREATE TABLE context_objects (
    -- Core MCP fields
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('task', 'code', 'chat')),
    version TEXT NOT NULL DEFAULT '1.0.0',
    timestamp INTEGER NOT NULL,
    data TEXT NOT NULL, -- JSON string containing the full object data

    -- Extended metadata fields
    session_id TEXT,
    project_id TEXT,
    owner TEXT,
    scope TEXT NOT NULL DEFAULT 'local' CHECK (scope IN ('local', 'shared')),
    permissions TEXT, -- JSON array of permission strings
    metadata TEXT, -- JSON object for key-value pairs
    tags TEXT, -- JSON array of tag strings

    -- Audit fields
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    created_by TEXT,
    updated_by TEXT,

    -- Foreign key constraints
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL
);

-- Create indexes for context_objects
CREATE INDEX idx_context_type ON context_objects(type);
CREATE INDEX idx_context_session_id ON context_objects(session_id);
CREATE INDEX idx_context_project_id ON context_objects(project_id);
CREATE INDEX idx_context_owner ON context_objects(owner);
CREATE INDEX idx_context_scope ON context_objects(scope);
CREATE INDEX idx_context_created_at ON context_objects(created_at);
CREATE INDEX idx_context_updated_at ON context_objects(updated_at);
CREATE INDEX idx_context_timestamp ON context_objects(timestamp);

-- Tasks table - specialized storage for task objects with queryable fields
CREATE TABLE tasks (
    -- Core identifiers
    id TEXT PRIMARY KEY,
    context_id TEXT NOT NULL UNIQUE, -- References context_objects.id

    -- Task-specific fields (denormalized for efficient querying)
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'blocked', 'cancelled')),
    priority TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),

    -- Hierarchical structure
    parent_id TEXT, -- References tasks.id for subtasks

    -- Assignment and estimation
    assignee TEXT,
    estimated_hours REAL CHECK (estimated_hours >= 0),
    actual_hours REAL CHECK (actual_hours >= 0),

    -- Dates
    due_date DATETIME,
    started_at DATETIME,
    completed_at DATETIME,

    -- Audit fields (duplicated from context_objects for performance)
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),

    -- Foreign key constraints
    FOREIGN KEY (context_id) REFERENCES context_objects(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE SET NULL
);

-- Create indexes for tasks
CREATE INDEX idx_tasks_context_id ON tasks(context_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX idx_tasks_assignee ON tasks(assignee);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX idx_tasks_status_priority ON tasks(status, priority);
CREATE INDEX idx_tasks_parent_status ON tasks(parent_id, status);

-- Task dependencies table - many-to-many relationship for task dependencies
CREATE TABLE task_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    depends_on_task_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),

    -- Foreign key constraints
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id) ON DELETE CASCADE,

    -- Ensure no duplicate dependencies
    UNIQUE(task_id, depends_on_task_id),

    -- Prevent self-dependencies
    CHECK (task_id != depends_on_task_id)
);

-- Create indexes for task_dependencies
CREATE INDEX idx_task_deps_task_id ON task_dependencies(task_id);
CREATE INDEX idx_task_deps_depends_on ON task_dependencies(depends_on_task_id);

-- Code references table - stores code file references for tasks
CREATE TABLE code_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    line_start INTEGER CHECK (line_start >= 0),
    line_end INTEGER CHECK (line_end >= line_start),
    description TEXT,
    commit_hash TEXT CHECK (length(commit_hash) = 40 OR commit_hash IS NULL),
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc')),

    -- Foreign key constraints
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Create indexes for code_references
CREATE INDEX idx_code_refs_task_id ON code_references(task_id);
CREATE INDEX idx_code_refs_file_path ON code_references(file_path);
CREATE INDEX idx_code_refs_commit_hash ON code_references(commit_hash);

-- Note: schema_migrations table is managed by the migration system

-- Create triggers to automatically update updated_at timestamps

-- Sessions trigger
CREATE TRIGGER trigger_sessions_updated_at
    AFTER UPDATE ON sessions
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE sessions SET updated_at = datetime('now', 'utc') WHERE id = NEW.id;
END;

-- Context objects trigger
CREATE TRIGGER trigger_context_objects_updated_at
    AFTER UPDATE ON context_objects
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE context_objects SET updated_at = datetime('now', 'utc') WHERE id = NEW.id;
END;

-- Tasks trigger
CREATE TRIGGER trigger_tasks_updated_at
    AFTER UPDATE ON tasks
    FOR EACH ROW
    WHEN NEW.updated_at = OLD.updated_at
BEGIN
    UPDATE tasks SET updated_at = datetime('now', 'utc') WHERE id = NEW.id;
END;

-- Trigger to automatically update task timestamps when status changes
CREATE TRIGGER trigger_tasks_status_change
    AFTER UPDATE OF status ON tasks
    FOR EACH ROW
    WHEN NEW.status != OLD.status
BEGIN
    UPDATE tasks SET
        started_at = CASE
            WHEN NEW.status = 'in_progress' AND OLD.status = 'pending' AND started_at IS NULL
            THEN datetime('now', 'utc')
            ELSE started_at
        END,
        completed_at = CASE
            WHEN NEW.status = 'completed' AND completed_at IS NULL
            THEN datetime('now', 'utc')
            ELSE completed_at
        END
    WHERE id = NEW.id;
END;
