package mtoc

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/xerrors"

	"coolstercodes/modules/modulir/mmarkdownext"
)

type header struct {
	level int
	id    string
	title string
}

// RenderFromHTML extracts a structure from the given HTML content and renders
// a corresponding table of contents as an HTML string.
func RenderFromHTML(content string) (string, error) {
	matches := headerRegexp.FindAllStringSubmatch(content, -1)
	headers := make([]*header, 0, len(matches))
	for _, match := range matches {
		level, err := strconv.Atoi(match[1])
		if err != nil {
			return "", xerrors.Errorf("error extracting header level: %w", err)
		}

		// Parse the header into an HTML doc
		doc, err := html.Parse(strings.NewReader(match[0]))
		if err != nil {
			return "", xerrors.Errorf("error extracting header level: %w", err)
		}

		// Extract the h* element
		parsedHeader := mmarkdownext.FindH(doc)
		// Do DFS to calculate the slug
		headerText := strings.TrimSpace(mmarkdownext.DFS(parsedHeader))
		slug := mmarkdownext.Slugify(headerText)
		headers = append(headers, &header{level, "#" + slug, headerText})
	}

	node := buildTree(headers)

	// Handle an article that doesn't have any TOC.
	if node == nil {
		return "", nil
	}

	return renderTree(node)
}

//
// Private
//

var headerRegexp = regexp.MustCompile(`<h([0-9]).* id="([^"]*)".*?>(<a.*?>)?(.*?)(</a>)?</h[0-9]>`)

func buildTree(headers []*header) *html.Node {
	if len(headers) < 1 {
		return nil
	}

	listNode := &html.Node{Data: "ol", Type: html.ElementNode}

	// keep a reference back to the top of the list
	topNode := listNode

	listItemNode := &html.Node{Data: "li", Type: html.ElementNode}
	listNode.AppendChild(listItemNode)

	// This basically helps us track whether we've insert multiple headers on
	// the same level in a row. If we did, we need to create a new list item
	// for each.
	needNewListNode := false

	var level int
	if len(headers) > 0 {
		level = headers[0].level
	}

	for _, header := range headers {
		if header.level > level {
			// indent

			// for each level indented, create a new nested list
			for range header.level - level {
				listNode = &html.Node{Data: "ol", Type: html.ElementNode}
				listItemNode.AppendChild(listNode)
			}

			needNewListNode = true

			level = header.level
		} else if header.level < level {
			// dedent

			// for each level outdented, move up two parents, one for list item
			// and one for list
			for range level - header.level {
				listItemNode = listNode.Parent
				listNode = listItemNode.Parent
			}

			level = header.level
		}

		if needNewListNode {
			listItemNode = &html.Node{Data: "li", Type: html.ElementNode}
			listNode.AppendChild(listItemNode)
		}

		contentNode := &html.Node{Data: header.title, Type: html.TextNode}

		linkNode := &html.Node{
			Data: "a",
			Attr: []html.Attribute{
				{Namespace: "", Key: "href", Val: header.id},
			},
			Type: html.ElementNode,
		}
		linkNode.AppendChild(contentNode)
		listItemNode.AppendChild(linkNode)

		needNewListNode = true
	}

	return topNode
}

func renderTree(node *html.Node) (string, error) {
	var b bytes.Buffer

	if err := html.Render(&b, node); err != nil {
		return "", xerrors.Errorf("error rendering HTML: %w", err)
	}

	return b.String(), nil
}
