package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Entity represents a memory context entity with atomic facts
type Entity struct {
	Name         string        `json:"name" validate:"required,min=1,max=200"`
	EntityType   string        `json:"entityType" validate:"required,min=1,max=100"`
	Observations []Observation `json:"observations"`
	CreatedAt    time.Time     `json:"createdAt"`
	LastModified time.Time     `json:"lastModified"`
}

// Observation represents an atomic fact about an entity
type Observation struct {
	ID        string    `json:"id" validate:"required,uuid"`
	Text      string    `json:"text" validate:"required,min=1,max=1000"`
	CreatedAt time.Time `json:"createdAt"`
	Source    string    `json:"source" validate:"required,min=1,max=100"`
}

// Relation represents a relationship between two entities
type Relation struct {
	ID           string    `json:"id" validate:"required,uuid"`
	From         string    `json:"from" validate:"required,min=1,max=200"`
	To           string    `json:"to" validate:"required,min=1,max=200"`
	RelationType string    `json:"relationType" validate:"required,min=1,max=100"`
	CreatedAt    time.Time `json:"createdAt"`
}

// RelationSet represents a collection of relations
type RelationSet struct {
	Relations []Relation `json:"relations"`
}

// Validate validates the Entity struct
func (e *Entity) Validate() error {
	// Set timestamps if not provided
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.LastModified.IsZero() {
		e.LastModified = time.Now()
	}

	// Validate each observation
	for i := range e.Observations {
		if err := e.Observations[i].Validate(); err != nil {
			return err
		}
	}

	return validate.Struct(e)
}

// Validate validates the Observation struct
func (o *Observation) Validate() error {
	// Generate ID if not provided
	if o.ID == "" {
		o.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}

	// Set default source if not provided
	if o.Source == "" {
		o.Source = "user_input"
	}

	return validate.Struct(o)
}

// Validate validates the Relation struct
func (r *Relation) Validate() error {
	// Generate ID if not provided
	if r.ID == "" {
		r.ID = uuid.New().String()
	}

	// Set timestamp if not provided
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now()
	}

	return validate.Struct(r)
}

// ToJSON marshals the Entity to JSON
func (e *Entity) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON unmarshals JSON data into the Entity
func (e *Entity) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// ToJSON marshals the RelationSet to JSON
func (rs *RelationSet) ToJSON() ([]byte, error) {
	return json.Marshal(rs)
}

// FromJSON unmarshals JSON data into the RelationSet
func (rs *RelationSet) FromJSON(data []byte) error {
	return json.Unmarshal(data, rs)
}

// Factory Functions

// NewEntity creates a new Entity with the given name and type
func NewEntity(name, entityType string) *Entity {
	entity := &Entity{
		Name:         name,
		EntityType:   entityType,
		Observations: make([]Observation, 0),
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}

	return entity
}

// NewObservation creates a new Observation with the given text
func NewObservation(text string) Observation {
	return Observation{
		ID:        uuid.New().String(),
		Text:      text,
		CreatedAt: time.Now(),
		Source:    "user_input",
	}
}

// NewRelation creates a new Relation between two entities
func NewRelation(from, to, relationType string) Relation {
	return Relation{
		ID:           uuid.New().String(),
		From:         from,
		To:           to,
		RelationType: relationType,
		CreatedAt:    time.Now(),
	}
}

// Entity Helper Methods

// AddObservation adds a new observation to the entity
func (e *Entity) AddObservation(text string) {
	observation := NewObservation(text)
	e.Observations = append(e.Observations, observation)
	e.LastModified = time.Now()
}

// AddObservationWithSource adds a new observation with a specific source
func (e *Entity) AddObservationWithSource(text, source string) {
	observation := NewObservation(text)
	observation.Source = source
	e.Observations = append(e.Observations, observation)
	e.LastModified = time.Now()
}

// RemoveObservation removes an observation by ID
func (e *Entity) RemoveObservation(observationID string) bool {
	for i, obs := range e.Observations {
		if obs.ID == observationID {
			e.Observations = append(e.Observations[:i], e.Observations[i+1:]...)
			e.LastModified = time.Now()
			return true
		}
	}
	return false
}

// GetObservationCount returns the number of observations
func (e *Entity) GetObservationCount() int {
	return len(e.Observations)
}

// SearchObservations returns observations containing the search text
func (e *Entity) SearchObservations(searchText string) []Observation {
	var results []Observation
	for _, obs := range e.Observations {
		if contains(obs.Text, searchText) {
			results = append(results, obs)
		}
	}
	return results
}

// RelationSet Helper Methods

// AddRelation adds a new relation to the set
func (rs *RelationSet) AddRelation(from, to, relationType string) {
	relation := NewRelation(from, to, relationType)
	rs.Relations = append(rs.Relations, relation)
}

// RemoveRelation removes a relation by ID
func (rs *RelationSet) RemoveRelation(relationID string) bool {
	for i, rel := range rs.Relations {
		if rel.ID == relationID {
			rs.Relations = append(rs.Relations[:i], rs.Relations[i+1:]...)
			return true
		}
	}
	return false
}

// GetRelationsByEntity returns all relations involving the specified entity
func (rs *RelationSet) GetRelationsByEntity(entityName string) []Relation {
	var results []Relation
	for _, rel := range rs.Relations {
		if rel.From == entityName || rel.To == entityName {
			results = append(results, rel)
		}
	}
	return results
}

// GetRelationsByType returns all relations of the specified type
func (rs *RelationSet) GetRelationsByType(relationType string) []Relation {
	var results []Relation
	for _, rel := range rs.Relations {
		if rel.RelationType == relationType {
			results = append(results, rel)
		}
	}
	return results
}

// Helper function for case-insensitive string search
func contains(text, searchText string) bool {
	// Simple case-insensitive contains check
	// Convert to lowercase for case-insensitive search
	textLower := strings.ToLower(text)
	searchLower := strings.ToLower(searchText)
	return strings.Contains(textLower, searchLower)
}
