package mmarkdownext

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestTransformCaptions(t *testing.T) {
	assert.Equal(t, `Paul is a <a href="https://linkedin.com" target="_blank" class="text-myblue underline">baller</a>`, transformCaption("Paul is a [baller](https://linkedin.com)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t, `Paul is a <a href="#baller" class="text-myblue">baller</a>`, transformCaption("Paul is a [baller](#baller)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t, `Paul is a <a href="/tags/baller" class="text-myblue">baller</a>`, transformCaption("Paul is a [baller](/tags/baller)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t, `Paul is a <a href="/baller" class="text-myblue">baller</a>`, transformCaption("Paul is a [baller](../baller/)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t, `Paul is a <a href="/content/images/hey/baller.txt" download class="text-myblue underline">baller</a>`, transformCaption("Paul is a [baller](./baller.txt)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t, `Paul is a <a href="/content/images/hey/baller.txt" download class="text-myblue underline">baller</a>`, transformCaption("Paul is a [baller](./baller.txt)", &RenderOptions{ImgDir: "/content/images/hey"}))

	assert.Equal(t,
		`<a href="https://linked.com" target="_blank" class="text-myblue underline">Paul</a> is <a href="#quite" class="text-myblue">quite</a> the <a href="/content/images/hey/baller.txt" download class="text-myblue underline">baller</a> around <a href="/these" class="text-myblue">these</a> <a href="/tags/parts" class="text-myblue">parts</a>`,
		transformCaption("[Paul](https://linked.com) is [quite](#quite) the [baller](./baller.txt) around [these](../these/) [parts](/tags/parts)", &RenderOptions{ImgDir: "/content/images/hey"}),
	)
}

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

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/img.jpg">
  <img src="/content/images/hey/img.jpg" />
</a>
`,
		must(transformImages(`![](./img.jpg)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/img.gif">
  <img src="/content/images/hey/img.gif" />
</a>
`,
		must(transformImages(`![](./img.gif)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/img.svg">
  <img src="/content/images/hey/img.svg" />
</a>
`,
		must(transformImages(`![](./img.svg)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	// Uppercase and mixed-case extensions
	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/Specs.JPG">
  <img src="/content/images/hey/Specs.JPG" />
</a>
`,
		must(transformImages(`![](./Specs.JPG)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/photo.PNG">
  <img src="/content/images/hey/photo.PNG" />
</a>
`,
		must(transformImages(`![](./photo.PNG)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a data-fancybox="gallery" href="/content/images/hey/anim.GIF">
  <img src="/content/images/hey/anim.GIF" />
</a>
`,
		must(transformImages(`![](./anim.GIF)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <a data-fancybox="gallery" href="/content/images/hey/Specs.JPG" data-caption="A spec sheet">
    <img src="/content/images/hey/Specs.JPG" />
  </a>
  <figcaption>A spec sheet</figcaption>
</figure>
`,
		must(transformImages(`![](./Specs.JPG)
*A spec sheet*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)
}

func TestTransformLinkedImages(t *testing.T) {
	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/image.png" />
  </a>
  <figcaption>some caption, but sometimes might not have a caption</figcaption>
</figure>
`,
		must(transformLinkedImages(`[![](./image.png)](https://google.com)
*some caption, but sometimes might not have a caption*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/image.png" />
  </a>
</figure>
`,
		must(transformLinkedImages(`[![](./image.png)](https://google.com)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	// Test with relative link
	assert.Equal(t, `
<figure class="text-center">
  <a href="/some/page">
    <img src="/content/images/hey/photo.jpg" />
  </a>
  <figcaption>A cool photo</figcaption>
</figure>
`,
		must(transformLinkedImages(`[![](./photo.jpg)](/some/page)
*A cool photo*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	// Test with caption containing markdown
	assert.Equal(t, `
<figure class="text-center">
  <a href="https://example.com" target="_blank">
    <img src="/content/images/hey/test.png" />
  </a>
  <figcaption>Check out this <a href="https://linked.com" target="_blank" class="text-myblue underline">cool link</a></figcaption>
</figure>
`,
		must(transformLinkedImages(`[![](./test.png)](https://example.com)
*Check out this [cool link](https://linked.com)*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	// Test different image formats
	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/image.jpeg" />
  </a>
</figure>
`,
		must(transformLinkedImages(`[![](./image.jpeg)](https://google.com)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/image.gif" />
  </a>
</figure>
`,
		must(transformLinkedImages(`[![](./image.gif)](https://google.com)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/image.svg" />
  </a>
</figure>
`,
		must(transformLinkedImages(`[![](./image.svg)](https://google.com)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	// Uppercase and mixed-case extensions
	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/Specs.JPG" />
  </a>
</figure>
`,
		must(transformLinkedImages(`[![](./Specs.JPG)](https://google.com)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <a href="https://google.com" target="_blank">
    <img src="/content/images/hey/photo.PNG" />
  </a>
  <figcaption>A nice photo</figcaption>
</figure>
`,
		must(transformLinkedImages(`[![](./photo.PNG)](https://google.com)
*A nice photo*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)
}

func TestTransformPDFs(t *testing.T) {
	assert.Equal(t, `
<iframe width="100%" height="800" src="/content/images/hey/pdf.pdf">
</iframe>
<figcaption class="text-center">A cool pdf</figcaption>
`,
		must(transformPDFs(`![](./pdf.pdf)
*A cool pdf*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<iframe width="100%" height="800" src="/content/images/hey/pdf.pdf">
</iframe>
`,
		must(transformPDFs(`![](./pdf.pdf)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)
}

func TestTransformVideos(t *testing.T) {
	assert.Equal(t, `
<figure class="text-center">
  <video controls>
    <source src="/content/images/hey/video.mp4" type="video/mp4">
    Your browser does not support the video tag.
  </video>
  <figcaption class="text-center">A dope video</figcaption>
</figure>
`,
		must(transformVideos(`![](./video.mp4)
*A dope video*`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<video controls>
  <source src="/content/images/hey/video.mp4" type="video/mp4">
  Your browser does not support the video tag.
</video>
`,
		must(transformVideos(`![](./video.mp4)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)
}

func TestTransformYouTubeVideos(t *testing.T) {
	assert.Equal(t, `
<div class="relative pb-[56.25%] h-0 overflow-hidden w-full">
	<iframe class="absolute w-full h-full top-0 left-0 border-0" src="https://www.youtube.com/embed/1ad5dq0Wi-c" referrerpolicy="strict-origin-when-cross-origin">
	</iframe>
</div>
`,
		must(transformYouTubeVideos(`![](https://youtu.be/1ad5dq0Wi-c)`, &RenderOptions{})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <div class="relative pb-[56.25%] h-0 overflow-hidden w-full">
    <iframe class="absolute w-full h-full top-0 left-0 border-0" src="https://www.youtube.com/embed/1ad5dq0Wi-c" referrerpolicy="strict-origin-when-cross-origin">
    </iframe>
  </div>
  <figcaption class="text-center">My first tutorial on this site!</figcaption>
</figure>
`,
		must(transformYouTubeVideos(`![](https://youtu.be/1ad5dq0Wi-c)
*My first tutorial on this site!*`, &RenderOptions{})),
	)

	assert.Equal(t, `
<div class="relative pb-[56.25%] h-0 overflow-hidden w-full">
	<iframe class="absolute w-full h-full top-0 left-0 border-0" src="https://www.youtube.com/embed/1ad5dq0Wi-c?start=42" referrerpolicy="strict-origin-when-cross-origin">
	</iframe>
</div>
`,
		must(transformYouTubeVideos(`![](https://youtu.be/1ad5dq0Wi-c?t=42)`, &RenderOptions{})),
	)

	assert.Equal(t, `
<figure class="text-center">
  <div class="relative pb-[56.25%] h-0 overflow-hidden w-full">
    <iframe class="absolute w-full h-full top-0 left-0 border-0" src="https://www.youtube.com/embed/1ad5dq0Wi-c?start=42" referrerpolicy="strict-origin-when-cross-origin">
    </iframe>
  </div>
  <figcaption class="text-center">Timestamped video with caption</figcaption>
</figure>
`,
		must(transformYouTubeVideos(`![](https://youtu.be/1ad5dq0Wi-c?t=42)
*Timestamped video with caption*`, &RenderOptions{})),
	)
}

func TestFiles(t *testing.T) {
	assert.Equal(t, `
<a href="/content/images/hey/file.txt" download>file.txt</a>
`,
		must(transformFiles(`[file.txt](./file.txt)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a href="/content/images/hey/file.csv" download>file.csv</a>
`,
		must(transformFiles(`[file.csv](./file.csv)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)

	assert.Equal(t, `
<a href="/content/images/hey/binary" download>binary of dreams</a>
`,
		must(transformFiles(`[binary of dreams](./binary)`, &RenderOptions{ImgDir: "/content/images/hey"})),
	)
}

func TestIsSlug(t *testing.T) {
	assert.True(t, isSlug("../georgia-tech-omscs-ai-for-robotics-review-cs-7638/"))
	assert.False(t, isSlug("../georgia-tech-omscs-ai-for-robotics-review-cs-7638"))
}

func TestGetArticleURL(t *testing.T) {
	assert.Equal(t, "/georgia-tech-omscs-ai-for-robotics-review-cs-7638", getArticleURL("../georgia-tech-omscs-ai-for-robotics-review-cs-7638/"))
}

func TestTransformHeadingLinks(t *testing.T) {
	assert.Equal(t, `<a href="#reinforcement-learning" class="no-underline">Reinforcement Learning</a>`,
		must(transformHeadingLinks(`<a href="#reinforcement-learning">Reinforcement Learning</a>`,
			nil,
		)),
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

	// Links that already have target="_blank" should not get duplicated
	assert.Equal(t,
		`<a href="https://example.com" target="_blank">Example</a>`,
		must(transformLinksToTargetBlank(
			`<a href="https://example.com" target="_blank">Example</a>`,
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

func TestSluggify(t *testing.T) {
	assert.Equal(t, "georgia-tech", Slugify("Georgia Tech"))
	assert.Equal(t, "ai", Slugify("AI"))
	assert.Equal(t, "ballin-it-up", Slugify("Ballin' it up"))
	assert.Equal(t, "hey", Slugify("Hey!"))
	assert.Equal(t, "introducing---ryan-denney", Slugify("Introducing - Ryan Denney"))
}

func TestHeaders(t *testing.T) {
	assert.Equal(t, `<h2 id="introducing-ryan-denney-ladies-and-gentlemen-test"><a href="#introducing-ryan-denney-ladies-and-gentlemen-test" class="no-underline">Introducing </a><a href="https://www.linkedin.com/in/ryan-denney-1418001b9/"><code>Ryan Denney</code> ladies and <code>gentlemen</code></a><a href="#introducing-ryan-denney-ladies-and-gentlemen-test" class="no-underline"> Test</a></h2>`,
		must(transformHeaders(`<h2>Introducing <a href="https://www.linkedin.com/in/ryan-denney-1418001b9/"><code>Ryan Denney</code> ladies and <code>gentlemen</code></a> Test</h2>`,
			nil,
		)),
	)

	assert.Equal(t, `<h2 id="and-another-test-hey"><a href="#and-another-test-hey" class="no-underline">And another </a><a href="https://google.com"><strong>test</strong></a><a href="#and-another-test-hey" class="no-underline"> hey</a></h2>`,
		must(transformHeaders(`<h2>And another <a href="https://google.com"><strong>test</strong></a> hey</h2>`,
			nil,
		)),
	)

	assert.Equal(t, `<h2 id="baller-extraordinaire"><a href="#baller-extraordinaire" class="no-underline"><code>Baller</code> Extraordinaire</a></h2>`,
		must(transformHeaders(`<h2><code>Baller</code> Extraordinaire</h2>`,
			nil,
		)),
	)
}
