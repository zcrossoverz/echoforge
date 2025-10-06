package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostType represents a post type definition entity
type PostType struct {
	ID                uuid.UUID       `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name              string          `json:"name" gorm:"type:varchar(50);not null;unique"`
	DisplayName       string          `json:"displayName" gorm:"type:varchar(100);not null"`
	Description       string          `json:"description,omitempty" gorm:"type:text"`
	FieldDefinitions  json.RawMessage `json:"fieldDefinitions" gorm:"type:jsonb;not null;default:'{}'"`
	IsActive          bool            `json:"isActive" gorm:"not null;default:true"`
	RequiresApproval  bool            `json:"requiresApproval" gorm:"not null;default:false"`
	AllowsScheduling  bool            `json:"allowsScheduling" gorm:"not null;default:true"`
	AllowsAttachments bool            `json:"allowsAttachments" gorm:"not null;default:true"`
	CreatedAt         time.Time       `json:"createdAt" gorm:"not null;default:now()"`
	UpdatedAt         time.Time       `json:"updatedAt" gorm:"not null;default:now()"`
}

// TableName specifies the table name for GORM
func (PostType) TableName() string {
	return "post_types"
}

// Validate checks if the post type entity is valid
func (pt *PostType) Validate() error {
	if pt.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(pt.Name) > 50 {
		return fmt.Errorf("name cannot exceed 50 characters")
	}

	// Validate name format (lowercase, alphanumeric + underscore only)
	for _, char := range pt.Name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("name must contain only lowercase letters, numbers, and underscores")
		}
	}

	if pt.DisplayName == "" {
		return fmt.Errorf("display name is required")
	}

	if len(pt.DisplayName) > 100 {
		return fmt.Errorf("display name cannot exceed 100 characters")
	}

	// Validate field definitions is valid JSON
	if len(pt.FieldDefinitions) > 0 {
		var fieldDefs map[string]interface{}
		if err := json.Unmarshal(pt.FieldDefinitions, &fieldDefs); err != nil {
			return fmt.Errorf("field definitions must be valid JSON: %w", err)
		}
	}

	return nil
}

// GetFieldDefinitions returns the field definitions as a map
func (pt *PostType) GetFieldDefinitions() (map[string]interface{}, error) {
	var fieldDefs map[string]interface{}
	if len(pt.FieldDefinitions) == 0 {
		return fieldDefs, nil
	}

	err := json.Unmarshal(pt.FieldDefinitions, &fieldDefs)
	return fieldDefs, err
}

// SetFieldDefinitions sets the field definitions from a map
func (pt *PostType) SetFieldDefinitions(fieldDefs map[string]interface{}) error {
	data, err := json.Marshal(fieldDefs)
	if err != nil {
		return fmt.Errorf("failed to marshal field definitions: %w", err)
	}

	pt.FieldDefinitions = data
	return nil
}

// ValidatePostMetadata validates post metadata against this post type's field definitions
func (pt *PostType) ValidatePostMetadata(metadata map[string]interface{}) error {
	fieldDefs, err := pt.GetFieldDefinitions()
	if err != nil {
		return fmt.Errorf("failed to get field definitions: %w", err)
	}

	// Validate each field in metadata against its definition
	for key, value := range metadata {
		fieldDef, exists := fieldDefs[key]
		if !exists {
			// Allow unknown fields for extensibility
			continue
		}

		if err := pt.validateFieldValue(key, value, fieldDef); err != nil {
			return err
		}
	}

	// Check for required fields
	for key, fieldDef := range fieldDefs {
		if pt.isFieldRequired(fieldDef) {
			if _, exists := metadata[key]; !exists {
				return fmt.Errorf("required field '%s' is missing", key)
			}
		}
	}

	return nil
}

// validateFieldValue validates a single field value against its definition
func (pt *PostType) validateFieldValue(key string, value interface{}, fieldDef interface{}) error {
	defMap, ok := fieldDef.(map[string]interface{})
	if !ok {
		return nil // Skip validation if definition is not a proper object
	}

	// Check type
	if expectedType, exists := defMap["type"]; exists {
		if !pt.isValueOfType(value, expectedType.(string)) {
			return fmt.Errorf("field '%s' must be of type %s", key, expectedType)
		}
	}

	// Check string length constraints
	if strValue, ok := value.(string); ok {
		if maxLength, exists := defMap["maxLength"]; exists {
			if maxLen, ok := maxLength.(float64); ok && len(strValue) > int(maxLen) {
				return fmt.Errorf("field '%s' exceeds maximum length of %d", key, int(maxLen))
			}
		}
		if minLength, exists := defMap["minLength"]; exists {
			if minLen, ok := minLength.(float64); ok && len(strValue) < int(minLen) {
				return fmt.Errorf("field '%s' is below minimum length of %d", key, int(minLen))
			}
		}
	}

	// Check numeric constraints
	if numValue, ok := value.(float64); ok {
		if maximum, exists := defMap["maximum"]; exists {
			if max, ok := maximum.(float64); ok && numValue > max {
				return fmt.Errorf("field '%s' exceeds maximum value of %g", key, max)
			}
		}
		if minimum, exists := defMap["minimum"]; exists {
			if min, ok := minimum.(float64); ok && numValue < min {
				return fmt.Errorf("field '%s' is below minimum value of %g", key, min)
			}
		}
	}

	// Check array constraints
	if arrValue, ok := value.([]interface{}); ok {
		if maxItems, exists := defMap["maxItems"]; exists {
			if maxI, ok := maxItems.(float64); ok && len(arrValue) > int(maxI) {
				return fmt.Errorf("field '%s' exceeds maximum items of %d", key, int(maxI))
			}
		}
		if minItems, exists := defMap["minItems"]; exists {
			if minI, ok := minItems.(float64); ok && len(arrValue) < int(minI) {
				return fmt.Errorf("field '%s' is below minimum items of %d", key, int(minI))
			}
		}
	}

	// Check enum constraints
	if enum, exists := defMap["enum"]; exists {
		if enumArray, ok := enum.([]interface{}); ok {
			found := false
			for _, enumValue := range enumArray {
				if value == enumValue {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("field '%s' must be one of the allowed values", key)
			}
		}
	}

	return nil
}

// isValueOfType checks if a value matches the expected JSON schema type
func (pt *PostType) isValueOfType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		if num, ok := value.(float64); ok {
			return num == float64(int(num)) // Check if it's a whole number
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "json":
		// JSON type allows any value
		return true
	case "date":
		// Date should be a string in ISO 8601 format
		if str, ok := value.(string); ok {
			_, err := time.Parse(time.RFC3339, str)
			return err == nil
		}
		return false
	default:
		return true // Unknown types are allowed
	}
}

// isFieldRequired checks if a field is marked as required
func (pt *PostType) isFieldRequired(fieldDef interface{}) bool {
	if defMap, ok := fieldDef.(map[string]interface{}); ok {
		if required, exists := defMap["required"]; exists {
			if req, ok := required.(bool); ok {
				return req
			}
		}
	}
	return false
}

// IsSystemType checks if this is a system-defined post type
func (pt *PostType) IsSystemType() bool {
	systemTypes := []string{"blog", "manga", "news"}
	for _, sysType := range systemTypes {
		if pt.Name == sysType {
			return true
		}
	}
	return false
}

// CanBeDeleted checks if this post type can be deleted
func (pt *PostType) CanBeDeleted() bool {
	// System types cannot be deleted
	return !pt.IsSystemType()
}

// BeforeCreate GORM hook called before creating a record
func (pt *PostType) BeforeCreate() error {
	if pt.ID == uuid.Nil {
		pt.ID = uuid.New()
	}
	pt.CreatedAt = time.Now()
	pt.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate GORM hook called before updating a record
func (pt *PostType) BeforeUpdate() error {
	pt.UpdatedAt = time.Now()
	return nil
}
