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
<figure class="text-center">
  <a data-fancybox="gallery" href="/content/images/hey/img.png" data-caption="some puppies">
    <img src="/content/images/hey/img.png" />
  </a>
  <figcaption>some puppies</figcaption>
</figure>
`,
		must(transformImages(`![](./img.png)
*some puppies*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/img.png">
  <img src="/content/images/hey/img.png" />
</a>
`,
		must(transformImages(`![](./img.png)`, &RenderOptions{ImgDir: "/content/images/hey"})),
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
}

func TestTransformHeaders(t *testing.T) {
	assert.Equal(t, `
<h2 id="n1-in-a-nutshell" class="link"><a href="#n1-in-a-nutshell">N+1 in a nutshell</a></h2>

Intro here.
`,
		must(transformHeaders(`
## N+1 in a nutshell

Intro here.
`,
			nil,
		)),
	)
}

func TestTransformLinksTargetBlank(t *testing.T) {
	assert.Equal(t,
		`<a href="https://example.com" target="_blank">Example</a>`+
			`<span class="hello">Hello</span>`,
		must(transformLinksToTargetBlank(
			`<a href="https://example.com">Example</a>`+
				`<span class="hello">Hello</span>`,
			&RenderOptions{},
		)),
	)

	// URLs that are relative should be left alone.
	assert.Equal(t,
		`<a href="/relative">Relative link</a>`,
		must(transformLinksToTargetBlank(
			`<a href="/relative">Relative link</a>`,
			&RenderOptions{},
		)),
	)
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}
	return v
}
