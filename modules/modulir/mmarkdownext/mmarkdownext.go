// Package mmarkdownext provides an extended version of Markdown that does
// several passes to add additional niceties like adding footnotes and allowing
// Go template helpers to be used..
package mmarkdownext

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/xerrors"
	"gopkg.in/russross/blackfriday.v2"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Public
//
//
//
//////////////////////////////////////////////////////////////////////////////

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

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Private
//
//
//
//////////////////////////////////////////////////////////////////////////////

// renderStack is the full set of functions that we'll run on an input string
// to get our fully rendered Markdown. This includes the rendering itself, but
// also a number of custom transformation options.
var renderStack = []func(string, *RenderOptions) (string, error){
	//
	// Pre-transformation functions
	//

	transformGoTemplate,
	transformHeaders,
	transformPDFs,
	transformImages,
	transformFiles,

	// The actual Blackfriday rendering
	func(source string, _ *RenderOptions) (string, error) {
		return string(blackfriday.Run([]byte(source))), nil
	},

	//
	// Post-transformation functions
	//

	// DEPRECATED: Find a different way to do this.
	transformCodeWithLanguagePrefix,

	transformFootnotes,

	transformLinksToTargetBlank,
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

var codeRE = regexp.MustCompile(`<code class="(\w+)">`)

func transformCodeWithLanguagePrefix(source string, _ *RenderOptions) (string, error) {
	return codeRE.ReplaceAllString(source, `<code class="language-$1">`), nil
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

var pdfRE = regexp.MustCompile(`(!\[\]\((.*).pdf\))(\n\*(.*)\*)?`)

func transformPDFs(source string, opts *RenderOptions) (string, error) {
	return pdfRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := figureRE.FindStringSubmatch(figure)
		if len(matches) != 5 {
			return figure
		}
		// Grab the pdf (it's the same every time)
		pdf := matches[2]
		if opts.ImgDir != "" {
			pdf = filepath.Join(opts.ImgDir, pdf)
		}

		// No caption option
		if matches[3] == "" {
			return fmt.Sprintf(pdfHTMLNoCaption, pdf)
		}

		// Grab the caption (only if 3rd arg isn't empty)
		caption := matches[4]
		return fmt.Sprintf(pdfHTMLCaption, pdf, caption)
	}), nil
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

// Let me break this regex down:
/*
	( - Starts first group
		!\[\] - Matches the "![]"
			Note: if we wanted to add alt text later, it would be !\[(.*)\]
		\( - matches first paren
		(.*) - matches everything until closing paren
		\) - matches closing paren
	) - Ends first group
	( - Starts second optional group
		\n - matches newline
		\* - matches first asterisk after newline
		(.*) - matches everything until next asterisk
		\* - matches last asterisk
	) - Ends second optional group
	? - Makes second group option (in case there is no caption given)
*/
var figureRE = regexp.MustCompile(`(!\[\]\((.*)\))(\n\*(.*)\*)?`)

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

const fileHTML = `
<a href="%s" download">%s</a>
`

var fileRE = regexp.MustCompile(`\[(.*)\]\(\./(.*)\)`)

func transformFiles(source string, opts *RenderOptions) (string, error) {
	return fileRE.ReplaceAllStringFunc(source, func(figure string) string {
		matches := fileRE.FindStringSubmatch(figure)
		if len(matches) != 3 {
			return figure
		}
		// Grab the file (it's the same every time)
		file := matches[2]
		if opts.ImgDir != "" {
			file = filepath.Join(opts.ImgDir, file)
		}
		// Grab the display name
		display := matches[1]

		return fmt.Sprintf(fileHTML, file, display)
	}), nil
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

const headerHTML = `
<h%v id="%s" class="link">
	<a href="#%s">%s</a>
</h%v>
`

// Matches the following:
//
//	# header
//
// For now, only match ## or more so as to remove code comments from
// matches. We need a better way of doing that though.
var headerRE = regexp.MustCompile(`(?m:^(#{2,})\s+(.*?)?$)`)

var slugRegexp = regexp.MustCompile(`[^\w\s-]`) // allows word chars, space, and hyphen

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRegexp.ReplaceAllString(s, "") // remove punctuation/symbols
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")  // replace spaces with hyphens
	s = strings.ReplaceAll(s, "--", "-") // collapse double hyphens (optional)
	return s
}

func transformHeaders(source string, _ *RenderOptions) (string, error) {
	source = headerRE.ReplaceAllStringFunc(source, func(header string) string {
		matches := headerRE.FindStringSubmatch(header)

		level := len(matches[1])
		title := matches[2]
		newID := slugify(title)

		return collapseHTML(fmt.Sprintf(headerHTML, level, newID, newID, title, level))
	})

	return source, nil
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
