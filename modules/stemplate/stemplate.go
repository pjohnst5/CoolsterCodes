package stemplate

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"math/rand/v2"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FuncMap is a set of helper functions to make available in templates for the
// project.
var FuncMap = template.FuncMap{
	"Downcase":             downcase,
	"FormatTimeLocal":      formatTimeLocal,
	"FormatTimeWithMinute": formatTimeWithMinute,
	"FormatTimeYearMonth":  formatTimeYearMonth,
	"InKM":                 inKM,
	"Mod":                  mod,
	"MonthName":            monthName,
	"NumberWithDelimiter":  numberWithDelimiter,
	"Pace":                 pace,
	"RandIntN":             rand.IntN,
	"RenderPublishingInfo": renderPublishingInfo,
	"Sub":                  sub,
	"ToStars":              toStars,
	"TimeInLocal":          timeInLocal,
	"URLBaseExt":           urlBaseExt,
	"URLBaseFile":          urlBaseFile,
}

// LocalLocation is the location to show times in which use FormatTimeLocal.
var LocalLocation *time.Location

func downcase(s string) string {
	return strings.ToLower(s)
}

func formatTimeLocal(t time.Time) string {
	if LocalLocation == nil {
		panic("stemplate.LocalLocation must be set")
	}

	return toNonBreakingWhitespace(t.In(LocalLocation).Format("January 2, 2006"))
}

func formatTimeYearMonth(t time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2006"))
}

func formatTimeWithMinute(t time.Time) string {
	return toNonBreakingWhitespace(t.Format("January 2, 2006 15:04"))
}

// This is a little tricky, but converts normal spaces to non-breaking spaces
// so that we can guarantee that certain strings will appear entirely on the
// same line. This is useful for a star count for example, because it's easy to
// misread a rating if it's broken up. See here for details:
//
// https://github.com/brandur/sorg/pull/60
func toNonBreakingWhitespace(str string) string {
	return strings.ReplaceAll(str, " ", " ")
}

func inKM(m float64) float64 {
	return m / 1000.0
}

func mod(i, j int) int {
	return i % j
}

func monthName(m time.Month) string {
	return m.String()
}

// Changes a number to a string and uses a separator for groups of three
// digits. For example, 1000 changes to "1,000".
func numberWithDelimiter(sep rune, n int) string {
	s := strconv.Itoa(n)

	startOffset := 0
	var buff bytes.Buffer

	if n < 0 {
		startOffset = 1
		buff.WriteByte('-')
	}

	l := len(s)

	commaIndex := 3 - ((l - startOffset) % 3)

	if commaIndex == 3 {
		commaIndex = 0
	}

	for i := startOffset; i < l; i++ {
		if commaIndex == 3 {
			buff.WriteRune(sep)
			commaIndex = 0
		}
		commaIndex++

		buff.WriteByte(s[i])
	}

	return buff.String()
}

// pace calculates the pace of a run in time per kilometer. This comes out as a
// "clock" time like 4:52 which translates to "4 minutes and 52 seconds" per
// kilometer.
func pace(distance float64, duration time.Duration) string {
	speed := duration.Seconds() / inKM(distance)
	minutes := int64(speed / 60.0)
	seconds := int64(speed) % 60
	return fmt.Sprintf("%v:%02d", minutes, seconds)
}

func renderPublishingInfo(info map[string]string) template.HTML {
	s := ""

	for k, v := range info {
		s += fmt.Sprintf("<p><strong>%s</strong><br>%s</p>", k, v)
	}

	return template.HTML(s)
}

// There is no "round" function built into Go :/.
func round(f float64) float64 {
	return math.Floor(f + .5)
}

func sub(i, j int) int {
	return i - j
}

func timeInLocal(t time.Time) time.Time {
	return t.In(LocalLocation)
}

func toStars(n int) string {
	var stars string
	for range n {
		stars += "★ "
	}
	return toNonBreakingWhitespace(stars)
}

// Gets the extension of a file at a URL. Also downcases said extension.
// Extension is returned _without_ a dot unlike `filepath.Ext`.
func urlBaseExt(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	ext := filepath.Ext(u.Path)
	if len(ext) > 0 {
		return strings.ToLower(ext[1:])
	}
	return ""
}

func urlBaseFile(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	return filepath.Base(u.Path)
}
