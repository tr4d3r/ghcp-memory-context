-- Add status column to users table
ALTER TABLE users ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended'));

-- Create index on status
CREATE INDEX idx_users_status ON users(status);
