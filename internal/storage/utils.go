package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// GenerateID generates a new UUID for entity IDs
func GenerateID() string {
	return uuid.New().String()
}

// NowUTC returns the current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// WithTimeout creates a context with the specified timeout
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// NullString converts a *string to sql.NullString
func NullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullTime converts a *time.Time to sql.NullTime
func NullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// NullInt64 converts a *int64 to sql.NullInt64
func NullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// NullBool converts a *bool to sql.NullBool
func NullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

// StringFromNull converts sql.NullString to *string
func StringFromNull(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// TimeFromNull converts sql.NullTime to *time.Time
func TimeFromNull(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// Int64FromNull converts sql.NullInt64 to *int64
func Int64FromNull(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

// BoolFromNull converts sql.NullBool to *bool
func BoolFromNull(nb sql.NullBool) *bool {
	if !nb.Valid {
		return nil
	}
	return &nb.Bool
}

// DefaultString returns the string value or a default if nil
func DefaultString(s *string, defaultValue string) string {
	if s == nil {
		return defaultValue
	}
	return *s
}

// DefaultInt returns the int value or a default if nil
func DefaultInt(i *int, defaultValue int) int {
	if i == nil {
		return defaultValue
	}
	return *i
}

// DefaultBool returns the bool value or a default if nil
func DefaultBool(b *bool, defaultValue bool) bool {
	if b == nil {
		return defaultValue
	}
	return *b
}

// DefaultTime returns the time value or a default if nil
func DefaultTime(t *time.Time, defaultValue time.Time) time.Time {
	if t == nil {
		return defaultValue
	}
	return *t
}
