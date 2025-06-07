package tag

import (
	"regexp"
	"strings"
)

func ToURL(tag string) string {
	// Convert to lowercase
	tag = strings.ToLower(tag)

	// Remove all non-word characters and replace with a dash
	re := regexp.MustCompile(`[\s\W-]+`)
	tag = re.ReplaceAllString(tag, "-")

	// Trim leading and trailing dashes
	tag = strings.Trim(tag, "-")

	return tag
}

// Given a list of articles, return a map of
/*
	{
		"georgia-tech": [article1, article2,...,articlen]
		"ai": [article3, article4,...,articlen]
	}
*/
