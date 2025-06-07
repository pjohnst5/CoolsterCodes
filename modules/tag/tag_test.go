package tag

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestNewTag(t *testing.T) {
	assert.Equal(t, "georgia-tech", ToURL("Georgia Tech"))
	assert.Equal(t, "ai", ToURL("AI"))
	assert.Equal(t, "ballin-it-up", ToURL("Ballin' it up"))
	assert.Equal(t, "hey", ToURL("Hey!"))
}
