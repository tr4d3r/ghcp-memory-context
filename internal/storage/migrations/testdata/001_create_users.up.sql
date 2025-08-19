-- Create users table for testing
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (datetime('now', 'utc'))
);

-- Create index on email
CREATE INDEX idx_users_email ON users(email);
