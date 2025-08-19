package migrations

import (
	"io/fs"
	"testing"
)

func TestEmbedFiles(t *testing.T) {
	// Test that the embedded filesystem contains our test migrations
	entries, err := fs.ReadDir(testMigrations, "testdata")
	if err != nil {
		t.Fatalf("Failed to read embedded testdata directory: %v", err)
	}

	expectedFiles := []string{
		"001_create_users.down.sql",
		"001_create_users.up.sql",
		"002_add_user_status.down.sql",
		"002_add_user_status.up.sql",
	}

	fileMap := make(map[string]bool)
	for _, entry := range entries {
		fileMap[entry.Name()] = true
	}

	for _, expected := range expectedFiles {
		if !fileMap[expected] {
			t.Errorf("Expected file %s not found in embedded filesystem", expected)
		}
	}

	// Test reading a specific file
	content, err := fs.ReadFile(testMigrations, "testdata/001_create_users.up.sql")
	if err != nil {
		t.Fatalf("Failed to read embedded file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Embedded file content is empty")
	}
}
