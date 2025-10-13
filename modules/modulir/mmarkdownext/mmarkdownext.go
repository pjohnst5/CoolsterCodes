// Package mmarkdownext provides an extended version of Markdown that does
// several passes to add additional niceties like adding footnotes and allowing
// Go template helpers to be used..
package mmarkdownext

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/net/html"
	"golang.org/x/xerrors"
	"gopkg.in/russross/blackfriday.v2"
)

// FuncMap is the map of helper functions that will be used when passing the
// Markdown through a Go template step.
var FuncMap = template.FuncMap{}

// RenderOptions describes a rendering operation to be customized.
type RenderOptions struct {
	// TemplateData is data injected while rendering Go templates.
	TemplateData interface{}

	// ImgDir is the path to the images
	ImgDir string
}

// Render a Markdown string to HTML while applying all custom project-specific
// filters including footnotes and stable header links.
func Render(s string, options *RenderOptions) (string, error) {
	var err error
	for _, f := range renderStack {
		s, err = f(s, options)
		if err != nil {
			return "", err
		}
	}
	return s, nil
}

// renderStack is the full set of functions that we'll run on an input string
// to get our fully rendered Markdown. This includes the rendering itself, but
// also a number of custom transformation options.
var renderStack = []func(string, *RenderOptions) (string, error){
	transformGoTemplate,
	transformImages,
	transformPDFs,
	transformVideos,
	transformFiles,

	// The actual Blackfriday rendering
	func(source string, _ *RenderOptions) (string, error) {
		return string(blackfriday.Run([]byte(source))), nil
	},

	replaceAll,

	transformHeadingLinks,

	transformHeaders,

	// DEPRECATED: Find a different way to do this.
	transformCodeWithLanguagePrefix,

	transformFootnotes,

	transformLinksToTargetBlank,
}

// Note that this should come early as we currently rely on a later step to
// give images a retina srcset.
func transformGoTemplate(source string, options *RenderOptions) (string, error) {
	// Skip this step if it doesn't look like there's any Go template code
	// contained in the source. (This may be a premature optimization.)
	if !strings.Contains(source, "{{") {
		return source, nil
	}

	tmpl, err := template.New("fmarkdownTemp").Funcs(FuncMap).Parse(source)
	if err != nil {
		return "", xerrors.Errorf("error parsing template: %w", err)
	}

	var templateData interface{}
	if options != nil {
		templateData = options.TemplateData
	}

	// Run the template to verify the output.
	var b bytes.Buffer
	err = tmpl.Execute(&b, templateData)
	if err != nil {
		return "", xerrors.Errorf("error executing template: %w", err)
	}

	// fmt.Printf("output in = %v ...\n", b.String())
	return b.String(), nil
}

const figureHTMLCaption = `
<figure class="text-center">
  <a data-fancybox="gallery" href="%s" data-caption="%s">
    <img src="%s" />
  </a>
  <figcaption>%s</figcaption>
</figure>
`

const figureHTMLNoCaption = `
<a data-fancybox="gallery" href="%s">
  <img src="%s" />
</a>
`

var figureRE = regexp.MustCompile(`(!\[\]\(([^)]+\.(?:png|jpg|jpeg|gif|svg))\))(\n\*(.*)\*)?`)

func transformImages(source string, opts *RenderOptions) (string, error) {
	return figureRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := figureRE.FindStringSubmatch(figure)
		if len(matches) != 5 {
			return figure
		}
		// Grab the image (it's the same every time)
		img := matches[2]
		if opts.ImgDir != "" {
			img = filepath.Join(opts.ImgDir, img)
		}

		// No caption option
		if matches[3] == "" {
			return fmt.Sprintf(figureHTMLNoCaption, img, img)
		}

		// Grab the caption (only if 3rd arg isn't empty)
		caption := matches[4]
		return fmt.Sprintf(figureHTMLCaption, img, caption, img, caption)
	}), nil
}

const pdfHTMLCaption = `
<iframe width="100%%" height="800" src="%s">
</iframe>
<figcaption class="text-center">%s</figcaption>
`

const pdfHTMLNoCaption = `
<iframe width="100%%" height="800" src="%s">
</iframe>
`

var pdfRE = regexp.MustCompile(`!\[\]\(([^)]+\.pdf)\)(\n\*(.*)\*)?`)

func transformPDFs(source string, opts *RenderOptions) (string, error) {
	return pdfRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := pdfRE.FindStringSubmatch(figure)
		if len(matches) != 4 {
			return figure
		}
		// Grab the pdf (it's the same every time)
		pdf := matches[1]
		if opts.ImgDir != "" {
			pdf = filepath.Join(opts.ImgDir, pdf)
		}

		// No caption option
		if matches[3] == "" {
			return fmt.Sprintf(pdfHTMLNoCaption, pdf)
		}

		// Grab the caption
		caption := matches[3]
		return fmt.Sprintf(pdfHTMLCaption, pdf, caption)
	}), nil
}

const videoHTMLCaption = `
<figure class="text-center">
  <video controls>
    <source src="%s" type="video/mp4">
    Your browser does not support the video tag.
  </video>
  <figcaption class="text-center">%s</figcaption>
</figure>
`

const videoHTMLNoCaption = `
<video controls>
  <source src="%s" type="video/mp4">
  Your browser does not support the video tag.
</video>
`

var videoRE = regexp.MustCompile(`!\[\]\(([^)]+\.mp4)\)(\n\*(.*)\*)?`)

func transformVideos(source string, opts *RenderOptions) (string, error) {
	return videoRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := videoRE.FindStringSubmatch(figure)
		if len(matches) != 4 {
			return figure
		}
		// Grab the video (it's the same every time)
		video := matches[1]
		if opts.ImgDir != "" {
			video = filepath.Join(opts.ImgDir, video)
		}

		// No caption option
		if matches[3] == "" {
			return fmt.Sprintf(videoHTMLNoCaption, video)
		}

		// Grab the caption (only if 3rd arg isn't empty)
		caption := matches[3]
		return fmt.Sprintf(videoHTMLCaption, video, caption)
	}), nil
}

const fileHTML = `
<a href="%s" download">%s</a>
`

const slugHTML = `
<a href="%s">%s</a>
`

var fileRE = regexp.MustCompile(`\[(.*)\]\((\.[^)]+)\)`)

func transformFiles(source string, opts *RenderOptions) (string, error) {
	return fileRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := fileRE.FindStringSubmatch(figure)
		if len(matches) != 3 {
			return figure
		}

		// Grab the display name
		display := matches[1]

		// Grab the file
		file := matches[2]

		// If it's a slug (like to an article directory) make it a url to that article
		if isSlug(file) {
			url := getArticleURL(file)
			return fmt.Sprintf(slugHTML, url, display)
		}

		// Otherwise, treat it as a downloadable file
		if opts.ImgDir != "" {
			file = filepath.Join(opts.ImgDir, file)
		}

		return fmt.Sprintf(fileHTML, file, display)
	}), nil
}

func isSlug(source string) bool {
	// If the path is something like "../georgia-tech-omscs-ai-for-robotics-review-cs-7638/" then it's a path to another article
	return strings.HasSuffix(source, "/")
}

func getArticleURL(input string) string {
	clean := path.Clean(input) // → "../georgia-tech-omscs-ai-for-robotics-review-cs-7638"
	clean = strings.TrimPrefix(clean, "..")
	return clean
}

// Look for any whitespace between HTML tags.
var whitespaceRE = regexp.MustCompile(`>\s+<`)

// Simply collapses certain HTML snippets by removing newlines and whitespace
// between tags. This is mainline used to make HTML snippets readable as
// constants, but then to make them fit a little more nicely into the rendered
// markup.
func collapseHTML(html string) string {
	html = strings.ReplaceAll(html, "\n", "")
	html = whitespaceRE.ReplaceAllString(html, "><")
	html = strings.TrimSpace(html)
	return html
}

var slugRegexp = regexp.MustCompile(`[^\w\s-]`) // allows word chars, space, and hyphen

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegexp.ReplaceAllString(s, "") // remove punctuation/symbols
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-") // replace spaces with hyphens
	return s
}

func replaceAll(source string, _ *RenderOptions) (string, error) {
	source = strings.ReplaceAll(source, "’", "'")
	source = strings.ReplaceAll(source, "–", "-")
	return source, nil
}

const headingHTML = `
<a href="#%s" class="no-underline">%s</a>
`

var headingRE = regexp.MustCompile(`<a href="#[^"]+"`)

func transformHeadingLinks(source string, _ *RenderOptions) (string, error) {
	return headingRE.ReplaceAllStringFunc(source, func(link string) string {
		return link + " class=\"no-underline\""
	}), nil
}

// The whole point of this function is to change something like this:
// "<h2>Introducing – <a href=\"https://www.linkedin.com/in/ryan-denney-1418001b9/\"><code>Ryan Denney</code> ladies and <code>gentlemen</code></a> Test</h2>"
// to this
// "<h2 id="introducing-ryan-denney-ladies-and-gentlemen-test"><a href="#introducing-ryan-denney-ladies-and-gentlemen-test">Introducing – <a href=\"https://www.linkedin.com/in/ryan-denney-1418001b9/\"><code>Ryan Denney</code> ladies and <code>gentlemen</code></a> Test</a></h2>"
// So basically, with the 'id' in it so people can click on it and it links to that exact header

var headerRE = regexp.MustCompile(`(?s)<h2>(.*?)</h2>|<h3>(.*?)</h3>|<h4>(.*?)</h4>|<h5>(.*?)</h5>|<h6>(.*?)</h6>`)

func transformHeaders(source string, _ *RenderOptions) (string, error) {
	source = headerRE.ReplaceAllStringFunc(source, func(header string) string {
		// Parse the header into an HTML doc
		doc, err := html.Parse(strings.NewReader(header))
		if err != nil {
			return source
		}

		// Extract the h* element
		parsedHeader := FindH(doc)

		// Do DFS to calculate the slug
		headerText := strings.TrimSpace(DFS(parsedHeader))
		slug := Slugify(headerText)
		setAttr(parsedHeader, "id", slug)

		// Add inner <a> tag with link to self!
		walk(parsedHeader)

		// Parse it back into HTML text
		htmlText, err := renderBackToHTML(parsedHeader)
		if err != nil {
			return source
		}

		return htmlText
	})

	return source, nil
}

// Find the first <h*> node.
func FindH(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && strings.HasPrefix(n.Data, "h") {
		if len(n.Data) == 2 && n.Data[1] >= '1' && n.Data[1] <= '6' {
			return n
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindH(c); found != nil {
			return found
		}
	}
	return nil
}

func DFS(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(DFS(c))
	}
	return sb.String()
}

func setAttr(n *html.Node, key, val string) {
	if n.Type != html.ElementNode {
		return
	}

	for i, a := range n.Attr {
		if a.Key == key {
			n.Attr[i].Val = val
			return
		}
	}

	// Not found → append
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func walk(n *html.Node) {
	if n.Type == html.ElementNode && len(n.Data) == 2 && n.Data[0] == 'h' && n.Data[1] >= '1' && n.Data[1] <= '6' {
		for _, a := range n.Attr {
			if a.Key == "id" {
				wrapWithID(n)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c)
	}
}

func wrapWithID(h *html.Node) {
	if h == nil || h.Type != html.ElementNode {
		return
	}

	// find id
	var id string
	for _, attr := range h.Attr {
		if attr.Key == "id" {
			id = attr.Val
			break
		}
	}
	if id == "" {
		return
	}

	// If no anchors anywhere, do single-wrap and return.
	if !containsLink(h) {
		one := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: "href", Val: "#" + id},
				{Key: "class", Val: "no-underline"},
			},
		}
		// move all children into the single anchor
		var children []*html.Node
		for c := h.FirstChild; c != nil; c = c.NextSibling {
			children = append(children, c)
		}
		for _, c := range children {
			h.RemoveChild(c)
			one.AppendChild(c)
		}
		h.AppendChild(one)
		return
	}

	// Otherwise: there are links somewhere in this heading.
	// We'll iterate *direct children* in order, grouping consecutive children that
	// do NOT contain anchors into runs, and wrap each run into its own self-link.
	// Any direct child that *is* an <a> OR contains descendant <a> will be preserved as-is
	// (after flushing the run before it).
	var children []*html.Node
	for c := h.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}

	flushRun := func(run []*html.Node) {
		if len(run) == 0 {
			return
		}
		self := &html.Node{
			Type: html.ElementNode,
			Data: "a",
			Attr: []html.Attribute{
				{Key: "href", Val: "#" + id},
				{Key: "class", Val: "no-underline"},
			},
		}
		for _, r := range run {
			// remove from h and append to self in same order
			h.RemoveChild(r)
			self.AppendChild(r)
		}
		h.AppendChild(self)
	}

	var run []*html.Node
	for _, c := range children {
		// Treat a child as a separator if:
		//  - it is an <a> element itself (direct anchor child), OR
		//  - it contains any descendant <a> (so we avoid creating nested anchors inside it)
		switch {
		case c.Type == html.ElementNode && strings.EqualFold(c.Data, "a"):
			// direct <a> child
			flushRun(run)
			run = nil
			h.RemoveChild(c)
			h.AppendChild(c)

		case containsLink(c):
			// contains descendant <a>
			flushRun(run)
			run = nil
			h.RemoveChild(c)
			h.AppendChild(c)

		default:
			// accumulate into a run to be wrapped
			run = append(run, c)
		}
	}
	// flush remaining run (trailing text etc)
	flushRun(run)
}

// recursively check if there is any <a> descendant.
func containsLink(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && strings.EqualFold(c.Data, "a") {
			return true
		}
		if containsLink(c) {
			return true
		}
	}
	return false
}

func renderBackToHTML(n *html.Node) (string, error) {
	var b bytes.Buffer
	err := html.Render(&b, n)
	if err != nil {
		return "", xerrors.Errorf("error rednering to html %+v", err)
	}

	renderedHTML := b.String()
	return renderedHTML, nil
}

var codeRE = regexp.MustCompile(`<code class="(\w+)">`)

func transformCodeWithLanguagePrefix(source string, _ *RenderOptions) (string, error) {
	return codeRE.ReplaceAllString(source, `<code class="language-$1">`), nil
}

// A layer that we wrap the entire footer section in for styling purposes.
const footerWrapper = `
<div class="footnotes">
  %s
</div>
`

// HTML for a footnote within the document.
const footnoteAnchorHTML = `
<sup id="footnote-%s">
  <a href="#footnote-%s-source">%s</a>
</sup>
`

// HTML for a reference to a footnote within the document.
//
// Make sure there's a single space before the <sup> because we're replacing
// one as part of our search.
const footnoteReferenceHTML = `
<sup id="footnote-%s-source">
  <a href="#footnote-%s">%s</a>
</sup>
`

// Look for the section the section at the bottom of the page that looks like
// <p>[1] (the paragraph tag is there because Markdown will have already
// wrapped it by this point).
var footerRE = regexp.MustCompile(`(?ms:^<p>\[\d+\].*)`)

// Look for a single footnote within the footer.
var footnoteRE = regexp.MustCompile(`\[(\d+)\](\s+.*)`)

// Note that this must be a post-transform filter. If it wasn't, our Markdown
// renderer would not render the Markdown inside the footnotes layer because it
// would already be wrapped in HTML.
func transformFootnotes(source string, _ *RenderOptions) (string, error) {
	footer := footerRE.FindString(source)

	if footer != "" {
		// remove the footer for now
		source = strings.Replace(source, footer, "", 1)

		footer = footnoteRE.ReplaceAllStringFunc(footer, func(footnote string) string {
			// first create a footnote with an anchor that links can target
			matches := footnoteRE.FindStringSubmatch(footnote)
			number := matches[1]

			anchor := fmt.Sprintf(footnoteAnchorHTML, number, number, number) + matches[2]

			// Then replace all references in the body to this footnote.
			//
			// Note the leading space before ` [%s]`. This is a little hacky,
			// but is there to try and ensure that we don't try to replace
			// strings that look like footnote references, but aren't.
			// `KEYS[1]` from `/redis-cluster` is an example of one of these
			// strings that might be a false positive.
			reference := fmt.Sprintf(footnoteReferenceHTML, number, number, number)

			source = strings.ReplaceAll(source,
				fmt.Sprintf(` [%s]`, number),
				" "+collapseHTML(reference))

			return collapseHTML(anchor)
		})

		// and wrap the whole footer section in a layer for styling
		footer = fmt.Sprintf(footerWrapper, footer)
		source += footer
	}

	return source, nil
}

// This just always transforms any "http*" links to blank targets to open in new tabs.
var absoluteLinkRE = regexp.MustCompile(`<a href="http[^"]+"`)

func transformLinksToTargetBlank(source string, _ *RenderOptions) (string, error) {
	return absoluteLinkRE.ReplaceAllStringFunc(source, func(link string) string {
		return link + " target=\"_blank\""
	}), nil
}
