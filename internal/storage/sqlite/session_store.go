package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tr4d3r/ghcp-memory-context/internal/storage"
)

// SessionStore implementation methods for Driver
func (d *Driver) CreateSession(ctx context.Context, session *storage.Session) error {
	if d.db == nil {
		return storage.NewStorageError("create_session", "session", "", storage.ErrDatabaseClosed)
	}

	// Serialize metadata to JSON
	var metadataJSON sql.NullString
	if session.Metadata != nil && len(session.Metadata) > 0 {
		data, err := json.Marshal(session.Metadata)
		if err != nil {
			return storage.NewStorageError("create_session", "session", session.ID, err)
		}
		metadataJSON = sql.NullString{String: string(data), Valid: true}
	}

	// Set timestamps if not provided
	now := time.Now()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = now
	}
	if session.LastAccessedAt.IsZero() {
		session.LastAccessedAt = now
	}

	query := `
		INSERT INTO sessions (
			id, user_id, project_id, name, description, metadata, active,
			created_at, updated_at, last_accessed_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := d.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		storage.NullStringPtr(session.ProjectID),
		storage.NullStringPtr(session.Name),
		storage.NullStringPtr(session.Description),
		metadataJSON,
		session.Active,
		session.CreatedAt,
		session.UpdatedAt,
		session.LastAccessedAt,
		storage.NullTimePtr(session.ExpiresAt),
	)

	if err != nil {
		return storage.NewStorageError("create_session", "session", session.ID, err)
	}

	return nil
}

func (d *Driver) GetSession(ctx context.Context, id string) (*storage.Session, error) {
	if d.db == nil {
		return nil, storage.NewStorageError("get_session", "session", id, storage.ErrDatabaseClosed)
	}

	query := `
		SELECT id, user_id, project_id, name, description, metadata, active,
			   created_at, updated_at, last_accessed_at, expires_at
		FROM sessions
		WHERE id = ?`

	row := d.db.QueryRowContext(ctx, query, id)

	return d.scanSession(row)
}

func (d *Driver) UpdateSession(ctx context.Context, session *storage.Session) error {
	if d.db == nil {
		return storage.NewStorageError("update_session", "session", session.ID, storage.ErrDatabaseClosed)
	}

	// Serialize metadata to JSON
	var metadataJSON sql.NullString
	if session.Metadata != nil && len(session.Metadata) > 0 {
		data, err := json.Marshal(session.Metadata)
		if err != nil {
			return storage.NewStorageError("update_session", "session", session.ID, err)
		}
		metadataJSON = sql.NullString{String: string(data), Valid: true}
	}

	// Update timestamp
	session.UpdatedAt = time.Now()

	query := `
		UPDATE sessions
		SET user_id = ?, project_id = ?, name = ?, description = ?, metadata = ?,
			active = ?, updated_at = ?, last_accessed_at = ?, expires_at = ?
		WHERE id = ?`

	result, err := d.db.ExecContext(ctx, query,
		session.UserID,
		storage.NullStringPtr(session.ProjectID),
		storage.NullStringPtr(session.Name),
		storage.NullStringPtr(session.Description),
		metadataJSON,
		session.Active,
		session.UpdatedAt,
		session.LastAccessedAt,
		storage.NullTimePtr(session.ExpiresAt),
		session.ID,
	)

	if err != nil {
		return storage.NewStorageError("update_session", "session", session.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError("update_session", "session", session.ID, err)
	}

	if rowsAffected == 0 {
		return storage.NewStorageError("update_session", "session", session.ID, storage.ErrNotFound)
	}

	return nil
}

func (d *Driver) DeleteSession(ctx context.Context, id string) error {
	if d.db == nil {
		return storage.NewStorageError("delete_session", "session", id, storage.ErrDatabaseClosed)
	}

	result, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return storage.NewStorageError("delete_session", "session", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return storage.NewStorageError("delete_session", "session", id, err)
	}

	if rowsAffected == 0 {
		return storage.NewStorageError("delete_session", "session", id, storage.ErrNotFound)
	}

	return nil
}

func (d *Driver) ListSessions(ctx context.Context, filter storage.SessionFilter) ([]*storage.Session, error) {
	if d.db == nil {
		return nil, storage.NewStorageError("list_sessions", "session", "", storage.ErrDatabaseClosed)
	}

	query, args := d.buildSessionQuery(filter)

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, storage.NewStorageError("list_sessions", "session", "", err)
	}
	defer rows.Close()

	var sessions []*storage.Session
	for rows.Next() {
		session, err := d.scanSession(rows)
		if err != nil {
			return nil, storage.NewStorageError("list_sessions", "session", "", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, storage.NewStorageError("list_sessions", "session", "", err)
	}

	return sessions, nil
}

func (d *Driver) CleanupExpiredSessions(ctx context.Context, olderThan time.Duration) error {
	if d.db == nil {
		return storage.NewStorageError("cleanup_sessions", "session", "", storage.ErrDatabaseClosed)
	}

	cutoff := time.Now().Add(-olderThan)

	query := `DELETE FROM sessions WHERE expires_at IS NOT NULL AND expires_at < ?`

	_, err := d.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return storage.NewStorageError("cleanup_sessions", "session", "", err)
	}

	return nil
}

// Helper methods

func (d *Driver) scanSession(scanner interface{}) (*storage.Session, error) {
	var s storage.Session
	var projectID, name, description sql.NullString
	var metadataJSON sql.NullString
	var expiresAt sql.NullTime

	var err error
	switch scanner := scanner.(type) {
	case *sql.Row:
		err = scanner.Scan(
			&s.ID, &s.UserID, &projectID, &name, &description, &metadataJSON,
			&s.Active, &s.CreatedAt, &s.UpdatedAt, &s.LastAccessedAt, &expiresAt,
		)
	case *sql.Rows:
		err = scanner.Scan(
			&s.ID, &s.UserID, &projectID, &name, &description, &metadataJSON,
			&s.Active, &s.CreatedAt, &s.UpdatedAt, &s.LastAccessedAt, &expiresAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	// Convert nullable fields
	s.ProjectID = projectID.String
	s.Name = name.String
	s.Description = description.String
	s.ExpiresAt = storage.TimePtrFromNull(expiresAt)

	// Deserialize metadata
	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &s.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &s, nil
}

func (d *Driver) buildSessionQuery(filter storage.SessionFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// Base query
	query := `
		SELECT id, user_id, project_id, name, description, metadata, active,
			   created_at, updated_at, last_accessed_at, expires_at
		FROM sessions`

	// Add WHERE conditions
	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.ProjectID != nil {
		conditions = append(conditions, "project_id = ?")
		args = append(args, *filter.ProjectID)
	}

	if filter.Active != nil {
		conditions = append(conditions, "active = ?")
		args = append(args, *filter.Active)
	}

	if filter.CreatedAfter != nil {
		conditions = append(conditions, "created_at > ?")
		args = append(args, *filter.CreatedAfter)
	}

	if filter.CreatedBefore != nil {
		conditions = append(conditions, "created_at < ?")
		args = append(args, *filter.CreatedBefore)
	}

	if filter.LastAccessedAfter != nil {
		conditions = append(conditions, "last_accessed_at > ?")
		args = append(args, *filter.LastAccessedAfter)
	}

	if filter.LastAccessedBefore != nil {
		conditions = append(conditions, "last_accessed_at < ?")
		args = append(args, *filter.LastAccessedBefore)
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY (default to created_at desc)
	query += " ORDER BY created_at DESC"

	// Add LIMIT/OFFSET
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	return query, args
}
