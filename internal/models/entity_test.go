package models

import (
	"testing"
)

func TestNewEntity(t *testing.T) {
	entity := NewEntity("project_standards", "guideline")

	if entity.Name != "project_standards" {
		t.Errorf("Expected name 'project_standards', got '%s'", entity.Name)
	}

	if entity.EntityType != "guideline" {
		t.Errorf("Expected entityType 'guideline', got '%s'", entity.EntityType)
	}

	if entity.GetObservationCount() != 0 {
		t.Errorf("Expected 0 observations, got %d", entity.GetObservationCount())
	}

	if entity.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestEntityAddObservation(t *testing.T) {
	entity := NewEntity("project_standards", "guideline")

	entity.AddObservation("use conventional commits")

	if entity.GetObservationCount() != 1 {
		t.Errorf("Expected 1 observation, got %d", entity.GetObservationCount())
	}

	obs := entity.Observations[0]
	if obs.Text != "use conventional commits" {
		t.Errorf("Expected observation text 'use conventional commits', got '%s'", obs.Text)
	}

	if obs.Source != "user_input" {
		t.Errorf("Expected default source 'user_input', got '%s'", obs.Source)
	}
}

func TestEntityAddObservationWithSource(t *testing.T) {
	entity := NewEntity("api_patterns", "pattern")

	entity.AddObservationWithSource("use REST endpoints", "code_analysis")

	obs := entity.Observations[0]
	if obs.Source != "code_analysis" {
		t.Errorf("Expected source 'code_analysis', got '%s'", obs.Source)
	}
}

func TestEntityRemoveObservation(t *testing.T) {
	entity := NewEntity("project_standards", "guideline")
	entity.AddObservation("use conventional commits")
	entity.AddObservation("format: type(scope): description")

	observationID := entity.Observations[0].ID
	removed := entity.RemoveObservation(observationID)

	if !removed {
		t.Error("Expected observation to be removed")
	}

	if entity.GetObservationCount() != 1 {
		t.Errorf("Expected 1 observation after removal, got %d", entity.GetObservationCount())
	}
}

func TestEntitySearchObservations(t *testing.T) {
	entity := NewEntity("project_standards", "guideline")
	entity.AddObservation("use conventional commits")
	entity.AddObservation("format: type(scope): description")
	entity.AddObservation("use REST API patterns")

	results := entity.SearchObservations("commit")

	if len(results) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(results))
	}

	if results[0].Text != "use conventional commits" {
		t.Errorf("Expected search result 'use conventional commits', got '%s'", results[0].Text)
	}
}

func TestEntityJSONMarshaling(t *testing.T) {
	entity := NewEntity("project_standards", "guideline")
	entity.AddObservation("use conventional commits")

	// Test ToJSON
	jsonData, err := entity.ToJSON()
	if err != nil {
		t.Fatalf("Failed to marshal entity to JSON: %v", err)
	}

	// Test FromJSON
	var newEntity Entity
	err = newEntity.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal entity from JSON: %v", err)
	}

	if newEntity.Name != entity.Name {
		t.Errorf("Expected name '%s' after JSON round-trip, got '%s'", entity.Name, newEntity.Name)
	}

	if newEntity.GetObservationCount() != entity.GetObservationCount() {
		t.Errorf("Expected %d observations after JSON round-trip, got %d",
			entity.GetObservationCount(), newEntity.GetObservationCount())
	}
}

func TestEntityValidation(t *testing.T) {
	// Test valid entity
	entity := NewEntity("project_standards", "guideline")
	err := entity.Validate()
	if err != nil {
		t.Errorf("Expected valid entity to pass validation, got error: %v", err)
	}

	// Test invalid entity (empty name)
	invalidEntity := &Entity{
		Name:       "",
		EntityType: "guideline",
	}
	err = invalidEntity.Validate()
	if err == nil {
		t.Error("Expected invalid entity (empty name) to fail validation")
	}
}

func TestNewRelation(t *testing.T) {
	relation := NewRelation("project", "standards", "follows")

	if relation.From != "project" {
		t.Errorf("Expected from 'project', got '%s'", relation.From)
	}

	if relation.To != "standards" {
		t.Errorf("Expected to 'standards', got '%s'", relation.To)
	}

	if relation.RelationType != "follows" {
		t.Errorf("Expected relationType 'follows', got '%s'", relation.RelationType)
	}

	if relation.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestRelationSetOperations(t *testing.T) {
	rs := &RelationSet{}

	// Add relations
	rs.AddRelation("project", "standards", "follows")
	rs.AddRelation("project", "patterns", "implements")
	rs.AddRelation("standards", "commit_format", "defines")

	if len(rs.Relations) != 3 {
		t.Errorf("Expected 3 relations, got %d", len(rs.Relations))
	}

	// Test GetRelationsByEntity
	projectRelations := rs.GetRelationsByEntity("project")
	if len(projectRelations) != 2 {
		t.Errorf("Expected 2 relations for 'project', got %d", len(projectRelations))
	}

	// Test GetRelationsByType
	followsRelations := rs.GetRelationsByType("follows")
	if len(followsRelations) != 1 {
		t.Errorf("Expected 1 'follows' relation, got %d", len(followsRelations))
	}

	// Test RemoveRelation
	relationID := rs.Relations[0].ID
	removed := rs.RemoveRelation(relationID)
	if !removed {
		t.Error("Expected relation to be removed")
	}

	if len(rs.Relations) != 2 {
		t.Errorf("Expected 2 relations after removal, got %d", len(rs.Relations))
	}
}

func TestRelationSetJSONMarshaling(t *testing.T) {
	rs := &RelationSet{}
	rs.AddRelation("project", "standards", "follows")

	// Test ToJSON
	jsonData, err := rs.ToJSON()
	if err != nil {
		t.Fatalf("Failed to marshal RelationSet to JSON: %v", err)
	}

	// Test FromJSON
	var newRS RelationSet
	err = newRS.FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to unmarshal RelationSet from JSON: %v", err)
	}

	if len(newRS.Relations) != len(rs.Relations) {
		t.Errorf("Expected %d relations after JSON round-trip, got %d",
			len(rs.Relations), len(newRS.Relations))
	}
}

func TestObservationValidation(t *testing.T) {
	// Test valid observation
	obs := NewObservation("use conventional commits")
	err := obs.Validate()
	if err != nil {
		t.Errorf("Expected valid observation to pass validation, got error: %v", err)
	}

	// Test observation with empty text
	invalidObs := Observation{
		Text: "",
	}
	err = invalidObs.Validate()
	if err == nil {
		t.Error("Expected invalid observation (empty text) to fail validation")
	}
}

func TestRelationValidation(t *testing.T) {
	// Test valid relation
	relation := NewRelation("project", "standards", "follows")
	err := relation.Validate()
	if err != nil {
		t.Errorf("Expected valid relation to pass validation, got error: %v", err)
	}

	// Test relation with empty from
	invalidRelation := Relation{
		From:         "",
		To:           "standards",
		RelationType: "follows",
	}
	err = invalidRelation.Validate()
	if err == nil {
		t.Error("Expected invalid relation (empty from) to fail validation")
	}
}
