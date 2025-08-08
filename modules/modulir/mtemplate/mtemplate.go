package mtemplate

import (
	"fmt"
	"html/template"
	"math"
	"regexp"
	"sort"
	"strings"
	texttemplate "text/template"
	"time"

	"golang.org/x/xerrors"
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

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap = template.FuncMap{
	"CollapseParagraphs":           CollapseParagraphs,
	"DistanceOfTimeInWords":        DistanceOfTimeInWords,
	"DistanceOfTimeInWordsFromNow": DistanceOfTimeInWordsFromNow,
	"FormatTime":                   FormatTime,
	"FormatTimeRFC3339UTC":         FormatTimeRFC3339UTC,
	"FormatTimeSimpleDate":         FormatTimeSimpleDate,
	"HTMLSafePassThrough":          HTMLSafePassThrough,
}

// CollapseParagraphs strips paragraph tags out of rendered HTML. Note that it
// does not handle HTML with any attributes, so is targeted mainly for use with
// HTML generated from Markdown.
func CollapseParagraphs(s string) string {
	sCollapsed := s
	sCollapsed = strings.ReplaceAll(sCollapsed, "<p>", "")
	sCollapsed = strings.ReplaceAll(sCollapsed, "</p>", "")
	return collapseHTML(sCollapsed)
}

// CombineFuncMaps combines a number of function maps into one. The combined
// version is a new function map so that none of the originals are tainted.
func CombineFuncMaps(funcMaps ...template.FuncMap) template.FuncMap {
	// Combine both sets of helpers into a single untainted function map.
	combined := make(template.FuncMap)

	for _, fm := range funcMaps {
		for k, v := range fm {
			if _, ok := combined[k]; ok {
				panic(xerrors.Errorf("duplicate function map key on combine: %s", k))
			}

			combined[k] = v
		}
	}

	return combined
}

// HTMLFuncMapToText transforms an HTML func map to a text func map.
func HTMLFuncMapToText(funcMap template.FuncMap) texttemplate.FuncMap {
	textFuncMap := make(texttemplate.FuncMap)

	for k, v := range funcMap {
		textFuncMap[k] = v
	}

	return textFuncMap
}

const (
	minutesInDay   = 24 * 60
	minutesInMonth = 30 * 24 * 60
	minutesInYear  = 365 * 24 * 60
)

// DistanceOfTimeInWords returns a string describing the relative time passed
// between two times.
func DistanceOfTimeInWords(to, from time.Time) string {
	d := from.Sub(to)

	minutes := int(round(d.Minutes()))

	switch {
	case minutes == 0:
		return "less than 1 minute"
	case minutes == 1:
		return fmt.Sprintf("%d minute", minutes)
	case minutes >= 1 && minutes <= 44:
		return fmt.Sprintf("%d minutes", minutes)
	case minutes >= 45 && minutes <= 89:
		return "about 1 hour"
	case minutes >= 90 && minutes <= minutesInDay-1:
		return fmt.Sprintf("about %d hours", int(round(d.Hours())))
	case minutes >= minutesInDay && minutes <= minutesInDay*2-1:
		return "about 1 day"
	case minutes >= 2520 && minutes <= minutesInMonth-1:
		return fmt.Sprintf("%d days", int(round(d.Hours()/24.0)))
	case minutes >= minutesInMonth && minutes <= minutesInMonth*2-1:
		return "about 1 month"
	case minutes >= minutesInMonth*2 && minutes <= minutesInYear-1:
		return fmt.Sprintf("%d months", int(round(d.Hours()/24.0/30.0)))
	case minutes >= minutesInYear && minutes <= minutesInYear+3*minutesInMonth-1:
		return "about 1 year"
	case minutes >= minutesInYear+3*minutesInMonth-1 && minutes <= minutesInYear+9*minutesInMonth-1:
		return "over 1 year"
	case minutes >= minutesInYear+9*minutesInMonth && minutes <= minutesInYear*2-1:
		return "almost 2 years"
	}

	return fmt.Sprintf("%d years", int(round(d.Hours()/24.0/365.0)))
}

// DistanceOfTimeInWordsFromNow returns a string describing the relative time
// passed between a time and the current moment.
func DistanceOfTimeInWordsFromNow(to time.Time) string {
	return DistanceOfTimeInWords(to, time.Now())
}

// HTMLSafePassThrough passes a string through to the final render. This is
// especially useful for code samples that contain Go template syntax which
// shouldn't be rendered.
func HTMLSafePassThrough(s string) template.HTML {
	return template.HTML(strings.TrimSpace(s))
}

// HTMLElement represents an HTML element that can be rendered.
type HTMLElement interface {
	render() template.HTML
}

// HTMLImage is a simple struct representing an HTML image to be rendered and
// some of the attributes it might have.
type HTMLImage struct {
	Src   string
	Alt   string
	Class string
}

// htmlElementRenderer is an internal representation of an HTML element to make
// building one with a set of properties easier.
type htmlElementRenderer struct {
	Name  string
	Attrs map[string]string
}

func (r *htmlElementRenderer) render() template.HTML {
	pairs := make([]string, 0, len(r.Attrs))
	for name, val := range r.Attrs {
		pairs = append(pairs, fmt.Sprintf(`%s="%s"`, name, val))
	}

	// Sort the outgoing names so that we have something stable to test against
	sort.Strings(pairs)

	return template.HTML(fmt.Sprintf(
		`<%s %s>`,
		r.Name,
		strings.Join(pairs, " "),
	))
}

func (img *HTMLImage) render() template.HTML {
	element := htmlElementRenderer{
		Name: "img",
		Attrs: map[string]string{
			"loading": "lazy",
			"src":     img.Src,
		},
	}

	if img.Alt != "" {
		element.Attrs["alt"] = img.Alt
	}

	if img.Class != "" {
		element.Attrs["class"] = img.Class
	}

	return element.render()
}

// FormatTime formats time according to the given format string.
func FormatTime(t time.Time, format string) string {
	return toNonBreakingWhitespace(t.Format(format))
}

// FormatTime formats time according to the given format string.
func FormatTimeRFC3339UTC(t time.Time) string {
	return toNonBreakingWhitespace(t.UTC().Format(time.RFC3339))
}

// FormatTimeSimpleDate formats time according to a relatively straightforward
// time format.
func FormatTimeSimpleDate(t time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2, 2006"))
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

// There is no "round" function built into Go :/.
func round(f float64) float64 {
	return math.Floor(f + .5)
}

func toNonBreakingWhitespace(str string) string {
	return strings.ReplaceAll(str, " ", " ")
}
