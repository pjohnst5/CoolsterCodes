package tag

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestNewTag(t *testing.T) {
	ta := NewTag("Georgia Tech")
	assert.Equal(t, "georgia-tech", ta.URL)
	assert.Equal(t, "Georgia Tech", ta.PrettyPrint)
}
