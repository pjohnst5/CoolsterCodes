package mtemplate

import (
	"html/template"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

var testTime time.Time

func init() {
	const longForm = "2006/01/02 15:04"
	var err error
	testTime, err = time.Parse(longForm, "2016/07/03 12:34")
	if err != nil {
		panic(err)
	}
}

func TestCollapseHTML(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>", collapseHTML(`
<p>
  <strong>strong</strong>
</p>`))
}

func TestCollapseParagraphs(t *testing.T) {
	assert.Equal(t, "<strong>strong</strong>", CollapseParagraphs(`
<p>
  <strong>strong</strong>
</p>
<p>
</p>`))
}

func TestCombineFuncMaps(t *testing.T) {
	fm1 := template.FuncMap{
		"CollapseParagraphs": CollapseParagraphs,
	}
	fm2 := template.FuncMap{
		"FormatTime": FormatTime,
	}
	fm3 := template.FuncMap{
		"FormatTimeRFC3339UTC": FormatTimeRFC3339UTC,
	}

	combined := CombineFuncMaps(fm1, fm2, fm3)

	{
		_, ok := combined["CollapseParagraphs"]
		assert.True(t, ok)
	}
	{
		_, ok := combined["FormatTime"]
		assert.True(t, ok)
	}
	{
		_, ok := combined["FormatTimeRFC3339UTC"]
		assert.True(t, ok)
	}
}

func TestCombineFuncMaps_Duplicate(t *testing.T) {
	fm1 := template.FuncMap{
		"CollapseParagraphs": CollapseParagraphs,
	}
	fm2 := template.FuncMap{
		"CollapseParagraphs": CollapseParagraphs,
	}

	assert.PanicsWithError(t,
		"duplicate function map key on combine: CollapseParagraphs", func() {
			_ = CombineFuncMaps(fm1, fm2)
		})
}

func TestHTMLFuncMapToText(t *testing.T) {
	fm := template.FuncMap{
		"FormatTimeRFC3339UTC": FormatTimeRFC3339UTC,
	}

	textFM := HTMLFuncMapToText(fm)

	{
		_, ok := textFM["FormatTimeRFC3339UTC"]
		assert.True(t, ok)
	}
}

func TestDistanceOfTimeInWords(t *testing.T) {
	to := time.Now()

	assert.Equal(t, "less than 1 minute",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-1s")), to))
	assert.Equal(t, "1 minute",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-1m")), to))
	assert.Equal(t, "8 minutes",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-8m")), to))
	assert.Equal(t, "about 1 hour",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-52m")), to))
	assert.Equal(t, "about 3 hours",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-3h")), to))
	assert.Equal(t, "about 1 day",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")), to))

	// note that parse only handles up to "h" units
	assert.Equal(t, "9 days",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*9), to))
	assert.Equal(t, "about 1 month",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*30), to))
	assert.Equal(t, "4 months",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*30*4), to))
	assert.Equal(t, "about 1 year",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*365), to))
	assert.Equal(t, "about 1 year",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365+2*30)), to))
	assert.Equal(t, "over 1 year",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365+3*30)), to))
	assert.Equal(t, "almost 2 years",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365+10*30)), to))
	assert.Equal(t, "2 years",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365*2)), to))
	assert.Equal(t, "3 years",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365*3)), to))
	assert.Equal(t, "10 years",
		DistanceOfTimeInWords(to.Add(mustParseDuration("-24h")*(365*10)), to))
}

func TestFormatTime(t *testing.T) {
	assert.Equal(t, "July 3, 2016 12:34", FormatTime(testTime, "January 2, 2006 15:04"))
}

func TestFormatTimeRFC3339UTC(t *testing.T) {
	assert.Equal(t, "2016-07-03T12:34:00Z", FormatTimeRFC3339UTC(testTime))
}

func TestFormatTimeSimpleDate(t *testing.T) {
	assert.Equal(t, "July 3, 2016", FormatTimeSimpleDate(testTime))
}

func TestHTMLImageRender(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		img := HTMLImage{Src: "src", Alt: "alt"}
		assert.Equal(
			t,
			`<img alt="alt" loading="lazy" src="src">`,
			string(img.render()),
		)
	})

	t.Run("NoSrcsetForSVG", func(t *testing.T) {
		img := HTMLImage{Src: "src.svg", Alt: "alt"}
		assert.Equal(
			t,
			`<img alt="alt" loading="lazy" src="src.svg">`,
			string(img.render()),
		)
	})

	t.Run("WithClass", func(t *testing.T) {
		img := HTMLImage{Src: "src", Alt: "alt", Class: "class"}
		assert.Equal(
			t,
			`<img alt="alt" class="class" loading="lazy" src="src">`,
			string(img.render()),
		)
	})
}

func TestHTMLSafePassThrough(t *testing.T) {
	assert.Equal(t, `{{print "x"}}`, string(HTMLSafePassThrough(`{{print "x"}}`)))
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

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return d
}
