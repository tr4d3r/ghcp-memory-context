-- Down migration for initial schema
-- This migration removes all tables and triggers created by 001_initial_schema.up.sql

-- Drop triggers first (they depend on tables)
DROP TRIGGER IF EXISTS trigger_tasks_status_change;
DROP TRIGGER IF EXISTS trigger_tasks_updated_at;
DROP TRIGGER IF EXISTS trigger_context_objects_updated_at;
DROP TRIGGER IF EXISTS trigger_sessions_updated_at;

-- Drop indexes (SQLite will drop them automatically with tables, but explicit is better)
DROP INDEX IF EXISTS idx_code_refs_commit_hash;
DROP INDEX IF EXISTS idx_code_refs_file_path;
DROP INDEX IF EXISTS idx_code_refs_task_id;

DROP INDEX IF EXISTS idx_task_deps_depends_on;
DROP INDEX IF EXISTS idx_task_deps_task_id;

DROP INDEX IF EXISTS idx_tasks_parent_status;
DROP INDEX IF EXISTS idx_tasks_status_priority;
DROP INDEX IF EXISTS idx_tasks_updated_at;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_assignee;
DROP INDEX IF EXISTS idx_tasks_parent_id;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_context_id;

DROP INDEX IF EXISTS idx_context_timestamp;
DROP INDEX IF EXISTS idx_context_updated_at;
DROP INDEX IF EXISTS idx_context_created_at;
DROP INDEX IF EXISTS idx_context_scope;
DROP INDEX IF EXISTS idx_context_owner;
DROP INDEX IF EXISTS idx_context_project_id;
DROP INDEX IF EXISTS idx_context_session_id;
DROP INDEX IF EXISTS idx_context_type;

DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_last_accessed;
DROP INDEX IF EXISTS idx_sessions_created_at;
DROP INDEX IF EXISTS idx_sessions_active;
DROP INDEX IF EXISTS idx_sessions_project_id;
DROP INDEX IF EXISTS idx_sessions_user_id;

-- Drop tables in reverse dependency order (children first, then parents)
DROP TABLE IF EXISTS code_references;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS context_objects;
DROP TABLE IF EXISTS sessions;

-- Drop schema migrations table last
DROP TABLE IF EXISTS schema_migrations;
