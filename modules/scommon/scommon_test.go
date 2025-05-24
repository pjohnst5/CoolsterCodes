package scommon

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestExtractSlug(t *testing.T) {
	// Article, fragment, etc.
	assert.Equal(t, "hello", ExtractSlug("hello.md"))

	// Newsletter
	assert.Equal(t, "001-hello", ExtractSlug("001-hello.md"))

	// With path included
	assert.Equal(t, "hello", ExtractSlug("path/to/hello.md"))
}
