package types

import (
	"fmt"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validation functions
	validate.RegisterValidation("uuid", validateUUID)
	validate.RegisterValidation("semver", validateSemVer)
}

// Validate performs validation on the BaseContext struct
func (bc *BaseContext) Validate() error {
	// Set timestamp if not provided
	if bc.Timestamp == 0 {
		bc.Timestamp = time.Now().Unix()
	}
	
	// Set created/updated timestamps if not provided
	now := time.Now()
	if bc.CreatedAt.IsZero() {
		bc.CreatedAt = now
	}
	if bc.UpdatedAt.IsZero() {
		bc.UpdatedAt = now
	}
	
	// Generate UUID if not provided
	if bc.ID == "" {
		bc.ID = uuid.New().String()
	}
	
	// Set default version if not provided
	if bc.Version == "" {
		bc.Version = "1.0.0"
	}
	
	// Set default scope if not provided
	if bc.Scope == "" {
		bc.Scope = ContextScopeLocal
	}
	
	return validate.Struct(bc)
}

// validateUUID checks if the value is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // Allow empty for omitempty fields
	}
	_, err := uuid.Parse(value)
	return err == nil
}

// validateSemVer checks if the value is a valid semantic version
func validateSemVer(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	
	// Basic semver pattern: MAJOR.MINOR.PATCH
	semverPattern := `^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`
	matched, err := regexp.MatchString(semverPattern, value)
	return err == nil && matched
}

// ValidateContextObject validates any context object that implements the ContextObject interface
func ValidateContextObject(obj ContextObject) error {
	if obj == nil {
		return fmt.Errorf("context object cannot be nil")
	}
	
	if obj.GetID() == "" {
		return fmt.Errorf("context object ID cannot be empty")
	}
	
	if obj.GetType() == "" {
		return fmt.Errorf("context object type cannot be empty")
	}
	
	if obj.GetVersion() == "" {
		return fmt.Errorf("context object version cannot be empty")
	}
	
	if obj.GetTimestamp() <= 0 {
		return fmt.Errorf("context object timestamp must be positive")
	}
	
	return obj.Validate()
}