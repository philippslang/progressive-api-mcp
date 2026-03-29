package unit_test

import (
	"testing"

	"github.com/philippslang/progressive-api-mcp/pkg/tools"
)

func BenchmarkPathAllowCheck10(b *testing.B) {
	allowed := make([]string, 10)
	for i := range allowed {
		allowed[i] = "/path/not/this"
	}
	allowed[9] = "/pets/{id}"
	b.ResetTimer()
	for range b.N {
		tools.IsPathPermitted("/pets/{id}", allowed)
	}
}

func BenchmarkPathAllowCheck50(b *testing.B) {
	allowed := make([]string, 50)
	for i := range allowed {
		allowed[i] = "/path/not/this"
	}
	allowed[49] = "/pets/{id}"
	b.ResetTimer()
	for range b.N {
		tools.IsPathPermitted("/pets/{id}", allowed)
	}
}

func BenchmarkPathAllowCheck100(b *testing.B) {
	allowed := make([]string, 100)
	for i := range allowed {
		allowed[i] = "/path/not/this"
	}
	allowed[99] = "/pets/{id}"
	b.ResetTimer()
	for range b.N {
		tools.IsPathPermitted("/pets/{id}", allowed)
	}
}
