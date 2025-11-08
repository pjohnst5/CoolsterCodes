package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/joeshaw/envdecode"
	"github.com/stretchr/testify/require"
)

func init() {
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding conf from env: %v", err)
		os.Exit(1)
	}
}

func TestExtCanonical(t *testing.T) {
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.JPG"))
	require.Equal(t, ".jpg", extCanonical("https://example.com/image.jpg?query"))
}

func TestExtImageTarget(t *testing.T) {
	require.Equal(t, ".jpg", extImageTarget(".jpg"))
	require.Equal(t, ".webp", extImageTarget(".heic"))
}

func TestLexicographicBase32(t *testing.T) {
	// Should only incorporate lower case characters.
	require.Equal(t, lexicographicBase32, strings.ToLower(lexicographicBase32))

	// All characters in the encoding set should be lexicographically ordered.
	{
		// This can be replaced with `strings.Clone` come Go 1.20.
		b := make([]byte, len(lexicographicBase32))
		copy(b, lexicographicBase32)

		slices.Sort(b)
		require.Equal(t, lexicographicBase32, string(b))
	}
}

func TestSimplifyMarkdownForSummary(t *testing.T) {
	require.Equal(t, "check that links are removed", simplifyMarkdownForSummary("check that [links](/link) are removed"))
	require.Equal(t, "double new lines are gone", simplifyMarkdownForSummary("double new\n\nlines are gone"))
	require.Equal(t, "single new lines are gone", simplifyMarkdownForSummary("single new\nlines are gone"))
	require.Equal(t, "space is trimmed", simplifyMarkdownForSummary(" space is trimmed "))
}

func TestTruncateString(t *testing.T) {
	require.Equal(t, "Short string unchanged.", truncateString("Short string unchanged.", 100))

	exactly100Length := strings.Repeat("s", 100)
	require.Equal(t, exactly100Length, truncateString(exactly100Length, 100))

	require.Equal(t,
		"This is a longer string that's going to need truncation and which will be truncated by ending it w …",
		truncateString("This is a longer string that's going to need truncation and which will be truncated by ending it with a space and an ellipsis.", 100),
	)
}

func TestInsertOrReplaceArticle(t *testing.T) {
	articles := []*Article{}

	// Test inserting a new article
	article1 := &Article{Slug: "article-1", Title: "Article 1"}
	insertOrReplaceArticle(&articles, article1)
	require.Len(t, articles, 1)
	require.Equal(t, "article-1", articles[0].Slug)
	require.Equal(t, "Article 1", articles[0].Title)

	// Test inserting another new article
	article2 := &Article{Slug: "article-2", Title: "Article 2"}
	insertOrReplaceArticle(&articles, article2)
	require.Len(t, articles, 2)
	require.Equal(t, "article-2", articles[1].Slug)

	// Test replacing an existing article
	article1Updated := &Article{Slug: "article-1", Title: "Article 1 Updated"}
	insertOrReplaceArticle(&articles, article1Updated)
	require.Len(t, articles, 2)
	require.Equal(t, "Article 1 Updated", articles[0].Title)
}

func TestInsertOrReplacePage(t *testing.T) {
	pages := []*Page{}

	// Test inserting a new page
	page1 := &Page{Slug: "page-1", Title: "Page 1"}
	insertOrReplacePage(&pages, page1)
	require.Len(t, pages, 1)
	require.Equal(t, "page-1", pages[0].Slug)
	require.Equal(t, "Page 1", pages[0].Title)

	// Test inserting another new page
	page2 := &Page{Slug: "page-2", Title: "Page 2"}
	insertOrReplacePage(&pages, page2)
	require.Len(t, pages, 2)
	require.Equal(t, "page-2", pages[1].Slug)

	// Test replacing an existing page
	page1Updated := &Page{Slug: "page-1", Title: "Page 1 Updated"}
	insertOrReplacePage(&pages, page1Updated)
	require.Len(t, pages, 2)
	require.Equal(t, "Page 1 Updated", pages[0].Title)
}
