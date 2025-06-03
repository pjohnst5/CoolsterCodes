package mmarkdownext

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestCollapseHTML(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>", collapseHTML(`
  <p>
  <strong>strong</strong>
</p>`))
}

func TestRender(t *testing.T) {
	assert.Equal(t, "<p><strong>strong</strong></p>\n", must(Render("**strong**", nil)))
}

func TestTransformCodeWithLanguagePrefix(t *testing.T) {
	assert.Equal(t,
		`<code class="language-ruby">`,
		must(transformCodeWithLanguagePrefix(`<code class="ruby">`, nil)),
	)
}

func TestTransformImages(t *testing.T) {
	assert.Equal(t, `
<figure>
  <a data-fancybox="gallery" href="/content/images/hey/img.png" data-caption="some puppies">
    <img src="/content/images/hey/img.png" />
  </a>
  <figcaption>some puppies</figcaption>
</figure>
`,
		must(transformImages(`![](/content/images/hey/img.png)
*some puppies*`, nil)),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/img.png">
  <img src="/content/images/hey/img.png" />
</a>
`,
		must(transformImages(`![](/content/images/hey/img.png)`, nil)),
	)
}

func TestTransformFootnotes(t *testing.T) {
	assert.Equal(t, `
<p>This is a reference <sup id="footnote-1-source"><a href="#footnote-1">1</a></sup>
to a footnote <sup id="footnote-2-source"><a href="#footnote-2">2</a></sup>.</p>

<p>Not footnote: KEYS[1].</p>


<div class="footnotes">
  <p><sup id="footnote-1"><a href="#footnote-1-source">1</a></sup> Footnote one.</p>

<p><sup id="footnote-2"><a href="#footnote-2-source">2</a></sup> Footnote two.</p>

</div>
`,
		must(transformFootnotes(`
<p>This is a reference [1]
to a footnote [2].</p>

<p>Not footnote: KEYS[1].</p>

<p>[1] Footnote one.</p>

<p>[2] Footnote two.</p>
`,
			nil,
		)),
	)

	// Without links
	assert.Equal(t, `
<p>This is a reference <sup><strong>1</strong></sup> to a footnote <sup><strong>2</strong></sup>.</p>

<p>Not footnote: KEYS[1].</p>


<div class="footnotes">
  <p><sup><strong>1</strong></sup> Footnote one.</p>

<p><sup><strong>2</strong></sup> Footnote two.</p>

</div>
`,
		must(transformFootnotes(`
<p>This is a reference [1] to a footnote [2].</p>

<p>Not footnote: KEYS[1].</p>

<p>[1] Footnote one.</p>

<p>[2] Footnote two.</p>
`,
			&RenderOptions{NoFootnoteLinks: true},
		)),
	)
}

func TestTransformHeaders(t *testing.T) {
	assert.Equal(t, `
<h2 id="intro" class="link"><a href="#intro">Introduction</a></h2>

Intro here.

<h2 id="section-1" class="link"><a href="#section-1">Body</a></h2>

<h3 id="article" class="link"><a href="#article">Article</a></h3>

Article one.

<h3 id="sub" class="link"><a href="#sub">Subsection</a></h3>

More content.

<h3 id="article-1" class="link"><a href="#article-1">Article</a></h3>

Article two.

<h3 id="section-5" class="link"><a href="#section-5">Subsection</a></h3>

More content.

<h2 id="conclusion" class="link"><a href="#conclusion">Conclusion</a></h2>

Conclusion.
`,
		must(transformHeaders(`
## Introduction (#intro)

Intro here.

## Body

### Article (#article)

Article one.

### Subsection (#sub)

More content.

### Article (#article)

Article two.

### Subsection

More content.

## Conclusion (#conclusion)

Conclusion.
`,
			nil,
		)),
	)

	assert.Equal(t, `
<h2>Introduction</h2>
`,
		must(transformHeaders(`
## Introduction (#intro)
`,
			&RenderOptions{NoHeaderLinks: true},
		)),
	)
}

func TestTransformImagesToAbsoluteURLs(t *testing.T) {
	// An image
	assert.Equal(t,
		`<img src="https://brandur.org/content/hello.jpg">`,
		must(transformImagesAndLinksToAbsoluteURLs(
			`<img src="/content/hello.jpg">`,
			&RenderOptions{AbsoluteURL: "https://brandur.org"},
		)),
	)

	// A link
	assert.Equal(t,
		`<a href="https://brandur.org/relative">Relative</a>`,
		must(transformImagesAndLinksToAbsoluteURLs(
			`<a href="/relative">Relative</a>`,
			&RenderOptions{AbsoluteURL: "https://brandur.org"},
		)),
	)

	// URLs that are already absolute are left alone.
	assert.Equal(t,
		`<img src="https://example.com/content/hello.jpg">`,
		must(transformImagesAndLinksToAbsoluteURLs(
			`<img src="https://example.com/content/hello.jpg">`,
			&RenderOptions{AbsoluteURL: "https://brandur.org"},
		)),
	)

	// Should pass through if options are nil.
	assert.Equal(t,
		`<img src="/content/hello.jpg">`,
		must(transformImagesAndLinksToAbsoluteURLs(
			`<img src="/content/hello.jpg">`,
			nil,
		)),
	)
}

func TestTransformLinksToNoFollow(t *testing.T) {
	assert.Equal(t,
		`<a href="https://example.com" rel="nofollow">Example</a>`+
			`<span class="hello">Hello</span>`,
		must(transformLinksToNoFollow(
			`<a href="https://example.com">Example</a>`+
				`<span class="hello">Hello</span>`,
			&RenderOptions{NoFollow: true},
		)),
	)

	// URLs that are relative should be left alone.
	assert.Equal(t,
		`<a href="/relative">Relative link</a>`,
		must(transformLinksToNoFollow(
			`<a href="/relative">Relative link</a>`,
			&RenderOptions{NoFollow: true},
		)),
	)

	// Should pass through if options are nil.
	assert.Equal(t,
		`<a href="https://example.com">Example</a>`,
		must(transformLinksToNoFollow(
			`<a href="https://example.com">Example</a>`,
			nil,
		)),
	)
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}
