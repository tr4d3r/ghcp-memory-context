package types

import (
	"encoding/json"
	"time"
)

// ContextType represents the type of context object
type ContextType string

const (
	ContextTypeTask ContextType = "task"
	ContextTypeCode ContextType = "code"
	ContextTypeChat ContextType = "chat"
)

// ContextScope defines whether context is local or shared
type ContextScope string

const (
	ContextScopeLocal  ContextScope = "local"
	ContextScopeShared ContextScope = "shared"
)

// BaseContext defines the core MCP-compliant context object structure
// All context objects must implement this base schema
type BaseContext struct {
	// Required MCP fields
	ID        string      `json:"id" validate:"required,uuid"`
	Type      ContextType `json:"type" validate:"required,oneof=task code chat"`
	Version   string      `json:"version" validate:"required,semver"`
	Timestamp int64       `json:"timestamp" validate:"required"`
	Data      interface{} `json:"data" validate:"required"`

	// Extended metadata fields for extensibility
	SessionID   string            `json:"session_id,omitempty" validate:"omitempty,uuid"`
	ProjectID   string            `json:"project_id,omitempty" validate:"omitempty,uuid"`
	Owner       string            `json:"owner,omitempty"`
	Scope       ContextScope      `json:"scope" validate:"required,oneof=local shared"`
	Permissions []string          `json:"permissions,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Tags        []string          `json:"tags,omitempty"`

	// Audit fields
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}

// ContextObject is the interface that all context types must implement
type ContextObject interface {
	GetID() string
	GetType() ContextType
	GetVersion() string
	GetTimestamp() int64
	GetData() interface{}
	Validate() error
	ToJSON() ([]byte, error)
	FromJSON([]byte) error
}

// Implement ContextObject interface for BaseContext
func (bc *BaseContext) GetID() string {
	return bc.ID
}

func (bc *BaseContext) GetType() ContextType {
	return bc.Type
}

func (bc *BaseContext) GetVersion() string {
	return bc.Version
}

func (bc *BaseContext) GetTimestamp() int64 {
	return bc.Timestamp
}

func (bc *BaseContext) GetData() interface{} {
	return bc.Data
}

func (bc *BaseContext) ToJSON() ([]byte, error) {
	return json.Marshal(bc)
}

func (bc *BaseContext) FromJSON(data []byte) error {
	return json.Unmarshal(data, bc)
}

// RelationshipType defines types of relationships between context objects
type RelationshipType string

const (
	RelationshipParentChild RelationshipType = "parent_child"
	RelationshipDependsOn   RelationshipType = "depends_on"
	RelationshipRelatedTo   RelationshipType = "related_to"
	RelationshipBlocks      RelationshipType = "blocks"
	RelationshipReferences  RelationshipType = "references"
)

// ContextRelationship defines relationships between context objects
type ContextRelationship struct {
	ID           string           `json:"id" validate:"required,uuid"`
	FromID       string           `json:"from_id" validate:"required,uuid"`
	ToID         string           `json:"to_id" validate:"required,uuid"`
	Type         RelationshipType `json:"type" validate:"required"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	CreatedBy    string           `json:"created_by,omitempty"`
}