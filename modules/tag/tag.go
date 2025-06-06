package tag

import (
	"strings"
)

type Tag struct {
	PrettyPrint string
	URL         string
}

// Assumes you give it the pretty print version
func NewTag(tag string) Tag {
	var t Tag
	t.PrettyPrint = tag
	// lowercase and replace spaces with dash
	lowerStr := strings.ToLower(tag)
	urlTag := strings.ReplaceAll(lowerStr, " ", "-")
	t.URL = urlTag
	return t
}

// Given a list of articles, return a map of
/*
	{
		"georgia-tech": [article1, article2,...,articlen]
		"ai": [article3, article4,...,articlen]
	}
*/
