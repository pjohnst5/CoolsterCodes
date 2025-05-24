package squantified

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCombineAuthors(t *testing.T) {
	assert.Equal(t,
		"Alex",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
		}),
	)

	assert.Equal(t,
		"Alex & Kate",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
		}),
	)

	assert.Equal(t,
		"Alex, Kate & Scan",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
			{Name: "Scan"},
		}),
	)

	assert.Equal(t,
		"Alex, Kate, Scan & Will",
		combineAuthors([]*ReadingAuthor{
			{Name: "Alex"},
			{Name: "Kate"},
			{Name: "Scan"},
			{Name: "Will"},
		}),
	)
}
