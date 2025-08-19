package migrations

import (
	"embed"
	"io/fs"
)

// EmbeddedMigrations contains the embedded migration files
//
//go:embed *.sql
var EmbeddedMigrations embed.FS

// GetEmbeddedFS returns the embedded filesystem for migrations
func GetEmbeddedFS() fs.FS {
	return EmbeddedMigrations
}

// GetMigrationsSubFS returns a sub-filesystem rooted at the migrations directory
func GetMigrationsSubFS() (fs.FS, error) {
	return fs.Sub(EmbeddedMigrations, ".")
}
