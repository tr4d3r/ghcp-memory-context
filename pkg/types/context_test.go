package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBaseContext_Validation(t *testing.T) {
	tests := []struct {
		name    string
		context BaseContext
		wantErr bool
	}{
		{
			name: "valid context with all fields",
			context: BaseContext{
				ID:        uuid.New().String(),
				Type:      ContextTypeTask,
				Version:   "1.0.0",
				Timestamp: time.Now().Unix(),
				Data:      map[string]interface{}{"test": "data"},
				Scope:     ContextScopeLocal,
			},
			wantErr: false,
		},
		{
			name: "valid context with minimal fields",
			context: BaseContext{
				Type: ContextTypeCode,
				Data: "some code context",
			},
			wantErr: false,
		},
		{
			name: "invalid context type",
			context: BaseContext{
				Type: "invalid_type",
				Data: "test",
			},
			wantErr: true,
		},
		{
			name: "missing data field",
			context: BaseContext{
				Type: ContextTypeChat,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.context.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseContext.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseContext_JSONSerialization(t *testing.T) {
	original := BaseContext{
		ID:        uuid.New().String(),
		Type:      ContextTypeTask,
		Version:   "1.0.0",
		Timestamp: time.Now().Unix(),
		Data:      map[string]interface{}{"title": "Test Task", "status": "pending"},
		Scope:     ContextScopeLocal,
		Tags:      []string{"test", "development"},
		Metadata:  map[string]string{"priority": "high"},
	}

	// Test serialization
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	// Test deserialization
	var restored BaseContext
	err = restored.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	// Verify key fields match
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: got %v, want %v", restored.ID, original.ID)
	}
	if restored.Type != original.Type {
		t.Errorf("Type mismatch: got %v, want %v", restored.Type, original.Type)
	}
	if restored.Version != original.Version {
		t.Errorf("Version mismatch: got %v, want %v", restored.Version, original.Version)
	}
}

func TestContextRelationship(t *testing.T) {
	relationship := ContextRelationship{
		ID:     uuid.New().String(),
		FromID: uuid.New().String(),
		ToID:   uuid.New().String(),
		Type:   RelationshipParentChild,
		Metadata: map[string]string{
			"description": "Parent task relationship",
		},
		CreatedAt: time.Now(),
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(relationship)
	if err != nil {
		t.Fatalf("Failed to serialize relationship: %v", err)
	}

	var restored ContextRelationship
	err = json.Unmarshal(jsonData, &restored)
	if err != nil {
		t.Fatalf("Failed to deserialize relationship: %v", err)
	}

	if restored.Type != relationship.Type {
		t.Errorf("Relationship type mismatch: got %v, want %v", restored.Type, relationship.Type)
	}
}