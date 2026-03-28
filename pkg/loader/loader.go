// Package loader parses and validates OpenAPI 3.x definition files.
package loader

import (
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
)

// Load reads the OpenAPI definition file at path, parses it with libopenapi,
// and returns the validated document. Returns an error if the file cannot be
// read or fails OpenAPI 3 structure validation.
func Load(path string) (libopenapi.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read OpenAPI definition %q: %w", path, err)
	}
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI definition %q: %w", path, err)
	}
	// Build the model to trigger full validation.
	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return nil, fmt.Errorf("invalid OpenAPI definition %q: %v", path, errs[0])
	}
	_ = model
	return doc, nil
}
