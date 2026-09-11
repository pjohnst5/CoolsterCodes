package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsMediaPathAndIsCodePath(t *testing.T) {
	mediaCases := []string{
		"content/articles/foo/bar.png",
		"content/images/logo.JPG",
		"content/pages/about/pic.jpeg",
		"content/articles/x/movie.mp4",
		"content/articles/x/doc.pdf",
	}
	for _, p := range mediaCases {
		require.Truef(t, isMediaPath(p), "expected media: %s", p)
		require.Falsef(t, isCodePath(p), "expected not code: %s", p)
	}

	codeCases := []string{
		"content/articles/foo/foo.md",
		"content/stylesheets/tailwind.css",
		"web/html/_nav.tmpl.html",
		"build.go",
		"package.json",
		"go.mod",
	}
	for _, p := range codeCases {
		require.Truef(t, isCodePath(p), "expected code: %s", p)
		require.Falsef(t, isMediaPath(p), "expected not media: %s", p)
	}

	// Belt-and-suspenders: even if a code extension were mistakenly added to
	// mediaExtensions, isMediaPath must still return false because
	// codeExtensions is consulted first.
	mediaExtensions[".md"] = true
	t.Cleanup(func() { delete(mediaExtensions, ".md") })
	require.False(t, isMediaPath("content/articles/foo/foo.md"),
		"code denylist must win over media allowlist")
}
