// Package registry holds all loaded API entries and provides lookup operations.
package registry

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/prograpimcp/prograpimcp/pkg/config"
	"github.com/prograpimcp/prograpimcp/pkg/loader"
	"github.com/prograpimcp/prograpimcp/pkg/validator"
)

// APIEntry is a fully loaded and validated API.
type APIEntry struct {
	// Name matches APIConfig.Name.
	Name string
	// Config is the original APIConfig for this entry.
	Config config.APIConfig
	// BaseURL is the resolved base URL (scheme + host + base_path).
	BaseURL string
	// Validator validates HTTP requests against this API's schema.
	Validator *validator.Validator
	// doc is the parsed libopenapi document.
	doc libopenapi.Document
}

// Document returns the parsed OpenAPI document for this API entry.
func (e APIEntry) Document() libopenapi.Document { return e.doc }

// Registry holds all loaded API entries.
type Registry struct {
	entries []APIEntry
	byName  map[string]int
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{byName: make(map[string]int)}
}

// Load parses, validates, and registers an APIConfig.
// Returns an error if the definition file is missing or fails OpenAPI validation.
func (r *Registry) Load(cfg config.APIConfig) error {
	doc, err := loader.Load(cfg.Definition)
	if err != nil {
		return fmt.Errorf("load API %q: %w", cfg.Name, err)
	}

	baseURL, err := resolveBaseURL(cfg, doc)
	if err != nil {
		return fmt.Errorf("resolve base URL for API %q: %w", cfg.Name, err)
	}

	v, err := validator.New(doc)
	if err != nil {
		return fmt.Errorf("build validator for API %q: %w", cfg.Name, err)
	}

	entry := APIEntry{
		Name:      cfg.Name,
		Config:    cfg,
		BaseURL:   baseURL,
		Validator: v,
		doc:       doc,
	}
	idx := len(r.entries)
	r.entries = append(r.entries, entry)
	r.byName[cfg.Name] = idx
	return nil
}

// Lookup returns the APIEntry for the given name (case-sensitive).
// Returns false if not found.
func (r *Registry) Lookup(name string) (APIEntry, bool) {
	idx, ok := r.byName[name]
	if !ok {
		return APIEntry{}, false
	}
	return r.entries[idx], true
}

// ListNames returns all registered API names in insertion order.
func (r *Registry) ListNames() []string {
	names := make([]string, len(r.entries))
	for i, e := range r.entries {
		names[i] = e.Name
	}
	return names
}

// Len returns the number of registered APIs.
func (r *Registry) Len() int { return len(r.entries) }

// resolveBaseURL computes the final base URL for an APIConfig.
// Priority: host+base_path > host only > servers[0].url from definition.
func resolveBaseURL(cfg config.APIConfig, doc libopenapi.Document) (string, error) {
	if cfg.Host != "" && cfg.BasePath != "" {
		base := strings.TrimRight(cfg.Host, "/")
		path := "/" + strings.TrimLeft(cfg.BasePath, "/")
		return base + path, nil
	}
	if cfg.Host != "" {
		return strings.TrimRight(cfg.Host, "/"), nil
	}

	// Fall back to servers[0].url in the OpenAPI document.
	model, errs := doc.BuildV3Model()
	if len(errs) > 0 {
		return "", fmt.Errorf("build model: %v", errs[0])
	}
	v3 := model.Model
	if v3.Servers == nil || len(v3.Servers) == 0 {
		return "", fmt.Errorf("no servers defined and no host configured")
	}
	serverURL := v3.Servers[0].URL
	if cfg.BasePath != "" {
		parsed, err := url.Parse(serverURL)
		if err != nil {
			return "", fmt.Errorf("parse server URL %q: %w", serverURL, err)
		}
		parsed.Path = "/" + strings.TrimLeft(cfg.BasePath, "/")
		return parsed.String(), nil
	}
	return strings.TrimRight(serverURL, "/"), nil
}
