package unit_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/prograpimcp/prograpimcp/pkg/loader"
	"github.com/prograpimcp/prograpimcp/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorValidate(t *testing.T) {
	doc, err := loader.Load(testdataPath("petstore.yaml"))
	require.NoError(t, err)
	v, err := validator.New(doc)
	require.NoError(t, err)

	t.Run("valid GET /pets", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/pets", nil)
		result := v.Validate(req)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("valid POST /pets", func(t *testing.T) {
		body := `{"name":"Fido","species":"dog"}`
		req, _ := http.NewRequest("POST", "/pets", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		result := v.Validate(req)
		assert.True(t, result.Valid)
	})

	t.Run("POST /pets missing required field", func(t *testing.T) {
		body := `{"species":"dog"}`
		req, _ := http.NewRequest("POST", "/pets", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		result := v.Validate(req)
		assert.False(t, result.Valid)
		require.NotEmpty(t, result.Errors)
		found := false
		for _, e := range result.Errors {
			if e.Field == "name" || e.Type == "MISSING_REQUIRED_FIELD" {
				found = true
			}
		}
		assert.True(t, found, "expected MISSING_REQUIRED_FIELD error for 'name'")
	})
}
