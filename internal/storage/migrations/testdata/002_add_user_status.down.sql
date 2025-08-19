-- Remove status column and its index
DROP INDEX IF EXISTS idx_users_status;

-- SQLite doesn't support DROP COLUMN, so we need to recreate the table
PRAGMA foreign_keys=off;

CREATE TABLE users_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc'))
);

INSERT INTO users_backup SELECT id, name, email, created_at FROM users;

DROP TABLE users;

ALTER TABLE users_backup RENAME TO users;

CREATE INDEX idx_users_email ON users(email);

PRAGMA foreign_keys=on;
