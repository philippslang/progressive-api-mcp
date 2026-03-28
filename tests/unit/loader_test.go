package unit_test

import (
	"testing"

	"github.com/prograpimcp/prograpimcp/pkg/loader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader(t *testing.T) {
	t.Run("valid OpenAPI 3.1 file", func(t *testing.T) {
		doc, err := loader.Load(testdataPath("petstore.yaml"))
		require.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loader.Load("/nonexistent/file.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "read")
	})

	t.Run("malformed OpenAPI", func(t *testing.T) {
		_, err := loader.Load(testdataPath("malformed.yaml"))
		require.Error(t, err)
	})
}
