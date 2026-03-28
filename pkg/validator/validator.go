// Package validator validates HTTP requests against OpenAPI 3.x schemas.
package validator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pb33f/libopenapi"
	libopenapiv "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/errors"
)

// ValidationError describes a single field-level schema violation.
type ValidationError struct {
	// Type is the error category.
	Type string
	// Field is the affected parameter or field name. Empty for PATH_NOT_FOUND.
	Field string
	// Message is a human-readable description suitable for agent self-correction.
	Message string
}

// Result holds the outcome of validating a single request.
type Result struct {
	// Valid is true if validation passed with no errors.
	Valid bool
	// Errors contains field-level validation errors when Valid is false.
	Errors []ValidationError
}

// Validator validates HTTP requests against a loaded OpenAPI document.
type Validator struct {
	v libopenapiv.Validator
}

// New creates a Validator from a libopenapi Document.
// Returns an error if the validator cannot be built from the document.
func New(doc libopenapi.Document) (*Validator, error) {
	v, errs := libopenapiv.NewValidator(doc)
	if len(errs) > 0 {
		return nil, fmt.Errorf("build validator: %v", errs[0])
	}
	return &Validator{v: v}, nil
}

// Validate checks whether the given http.Request conforms to the OpenAPI schema.
// Returns a Result with Valid=true and empty Errors on success.
func (val *Validator) Validate(r *http.Request) Result {
	valid, validationErrors := val.v.ValidateHttpRequest(r)
	if valid {
		return Result{Valid: true}
	}

	errs := make([]ValidationError, 0, len(validationErrors))
	for _, ve := range validationErrors {
		errs = append(errs, ValidationError{
			Type:    classifyError(ve),
			Field:   extractField(ve),
			Message: ve.Message,
		})
	}
	return Result{Valid: false, Errors: errs}
}

func classifyError(ve *errors.ValidationError) string {
	// Check schema-level errors first (most specific).
	for _, sve := range ve.SchemaValidationErrors {
		reason := strings.ToLower(sve.Reason)
		if strings.Contains(reason, "missing property") {
			return "MISSING_REQUIRED_FIELD"
		}
		if strings.Contains(reason, "additional") {
			return "ADDITIONAL_PROPERTY"
		}
		if strings.Contains(reason, "type") {
			return "INVALID_FIELD_TYPE"
		}
	}

	// Fall back to top-level message classification.
	msg := strings.ToLower(ve.Message)
	switch {
	case strings.Contains(msg, "required"):
		if strings.Contains(msg, "parameter") {
			return "MISSING_REQUIRED_PARAM"
		}
		return "MISSING_REQUIRED_FIELD"
	case strings.Contains(msg, "type"):
		if strings.Contains(msg, "parameter") {
			return "INVALID_PARAM_TYPE"
		}
		return "INVALID_FIELD_TYPE"
	case strings.Contains(msg, "additional"):
		return "ADDITIONAL_PROPERTY"
	default:
		return "SCHEMA_VIOLATION"
	}
}

func extractField(ve *errors.ValidationError) string {
	// Try to extract field name from schema validation errors.
	for _, sve := range ve.SchemaValidationErrors {
		reason := sve.Reason
		// Reason like "missing property 'name'" => extract 'name'.
		if idx := strings.Index(reason, "'"); idx >= 0 {
			end := strings.Index(reason[idx+1:], "'")
			if end >= 0 {
				return reason[idx+1 : idx+1+end]
			}
		}
		// Location like "/required" or "/properties/name" => extract last segment.
		loc := sve.Location
		if loc != "" && loc != "/required" {
			if idx := strings.LastIndex(loc, "/"); idx >= 0 {
				return loc[idx+1:]
			}
		}
	}
	return ""
}
