package main

import (
	"context"
	"html/template"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	"golang.org/x/xerrors"

	"github.com/brandur/sorg/modules/modulir"
	"github.com/brandur/sorg/modules/modulir/mfile"
	"github.com/brandur/sorg/modules/modulir/mmarkdownext"
	"github.com/brandur/sorg/modules/modulir/mtemplate"
	"github.com/brandur/sorg/modules/modulir/mtoc"
	"github.com/brandur/sorg/modules/modulir/mtoml"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/stemplate"
)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Constants
//
//
//
//////////////////////////////////////////////////////////////////////////////

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Variables
//
//
//
//////////////////////////////////////////////////////////////////////////////

// These are all objects that are persisted between build loops so that if
// necessary we can rebuild jobs that depend on them like index pages without
// reparsing all the source material. In each case we try to only reparse the
// sources if those source files actually changed.
var (
	articles     []*Article
	dependencies = NewDependencyRegistry()
	pages        = make(map[string]*Page)
)

// Time zone to show articles publishing times in.
var localLocation = mustLocation("America/Denver")

// List of common build dependencies, a change in any of which will trigger a
// rebuild on everything: partial html, JavaScripts, and stylesheets. Even
// though some of those changes will false positives, these sources are
// pervasive enough, and changes infrequent enough, that it's worth the
// tradeoff. This variable is a global because so many render functions access
// it.
var universalSources []string

var validate = validator.New()

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Init
//
//
//
//////////////////////////////////////////////////////////////////////////////

func init() {
	mmarkdownext.FuncMap = scommon.TextTemplateFuncMap
	stemplate.LocalLocation = localLocation
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Build function
//
//
//
//////////////////////////////////////////////////////////////////////////////

func build(c *modulir.Context) []error {
	//
	// PHASE 0: Setup
	//
	// (No jobs should be enqueued here.)
	//

	c.Log.Debugf("Running build loop")

	// This is where we stored "versioned" content like compiled JS and CSS.
	// These content have a release number that we can increment and by
	// extension quickly invalidate.
	versionedContentDir := path.Join(c.TargetDir, "content", Release)

	// A set of source paths that rebuild everything when any one of them
	// changes. These are dependencies that are included in more or less
	// everything: common partial html, JavaScript sources, and stylesheet
	// sources.
	universalSources = nil

	// Generate a set of JavaScript sources to add to universal sources.
	{
		javaScriptSources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/javascripts",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}
		universalSources = append(universalSources, javaScriptSources...)
	}

	// Generate a list of partial html to add to universal sources.
	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/html",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}

		var partialHTML []string
		for _, source := range sources {
			if strings.HasPrefix(filepath.Base(source), "_") {
				partialHTML = append(partialHTML, source)
			}
		}

		universalSources = append(universalSources, partialHTML...)
	}

	// Generate a set of stylesheet sources to add to universal sources.
	{
		stylesheetSources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/stylesheets",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}
		universalSources = append(universalSources, stylesheetSources...)
	}

	//
	// PHASE 1
	//
	// The build is broken into phases because some jobs depend on jobs that
	// ran before them. For example, we need to parse all our article metadata
	// before we can create an article index and render the home page (which
	// contains a short list of articles).
	//
	// After each phase, we call `Wait` on our context which will wait for the
	// worker pool to finish all its current work and restart it to accept new
	// jobs after it has.
	//
	// The general rule is to make sure that work is done as early as it
	// possibly can be. e.g. Jobs with no dependencies should always run in
	// phase 1. Try to make sure that as few phases as necessary.
	//

	ctx := context.Background()

	//
	// Common directories
	//
	// Create these outside of the job system because jobs below may depend on
	// their existence.
	//

	{
		commonDirs := []string{
			c.TargetDir + "/tags",
			scommon.TempDir,
			versionedContentDir,
		}
		for _, dir := range commonDirs {
			err := mfile.EnsureDir(c, dir)
			if err != nil {
				return []error{nil}
			}
		}
	}

	//
	// Symlinks
	//

	{
		commonSymlinks := [][2]string{
			{c.SourceDir + "/content/images", c.TargetDir + "/content/images"},
			{c.SourceDir + "/content/javascripts", versionedContentDir + "/javascripts"},
			{c.SourceDir + "/content/stylesheets", versionedContentDir + "/stylesheets"},
		}
		for _, link := range commonSymlinks {
			err := mfile.EnsureSymlink(c, link[0], link[1])
			if err != nil {
				return []error{nil}
			}
		}
	}

	//
	// Articles
	//

	var articlesChanged bool
	var articlesMu sync.Mutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/articles", nil)
		if err != nil {
			return []error{err}
		}

		for _, s := range sources {
			source := s

			name := "article: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderArticle(ctx, c, source,
					&articles, &articlesChanged, &articlesMu)
			})
		}
	}

	//
	// Pages (render each view)
	//

	var pagesMu sync.RWMutex

	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/html/pages", &mfile.ReadDirOptions{RecurseDirs: true})
		if err != nil {
			return []error{err}
		}

		for _, s := range sources {
			source := s

			name := "page: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderPage(ctx, c, source, pages, &pagesMu)
			})
		}
	}

	//
	//
	//
	// PHASE 2
	//
	//
	//

	if errors := c.Wait(); errors != nil {
		c.Log.Errorf("Cancelling next phase due to build errors")
		return errors
	}

	// Various sorts for anything that might need it.
	//
	// Some slices are sorted above when they're read in so that they can be
	// compared against a current version.
	{
		slices.SortFunc(articles, func(a, b *Article) int { return b.PublishedAt.Compare(a.PublishedAt) })
	}

	//
	// Home
	//

	{
		tagMap := getTagMap(articles)
		c.AddJob("home", func() (bool, error) {
			return renderHome(ctx, c, articles,
				articlesChanged, tagMap)
		})
	}

	//
	// Tags
	//
	{
		tagMap := getTagMap(articles)
		for tag, articles := range tagMap {
			c.AddJob(tag, func() (bool, error) {
				return renderTag(ctx, c,
					tag,
					articles,
					articlesChanged)
			})
		}
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Types
//
//
//
//////////////////////////////////////////////////////////////////////////////

// Article represents an article to be rendered.
type Article struct {
	// Attributions are any attributions for content that may be included in
	// the article (like an image in the header for example).
	Attributions template.HTML `toml:"attributions,omitempty"`

	// Content is the HTML content of the article. It isn't included as TOML
	// frontmatter, and is rather split out of an article's Markdown file,
	// rendered, and then added separately.
	Content template.HTML `toml:"-"`

	// Footnotes are HTML footnotes extracted from content.
	Footnotes template.HTML `toml:"-"`

	// HNLink is an optional link to comments on Hacker News.
	HNLink string `toml:"hn_link,omitempty"`

	// Hook is a leading sentence or two to succinctly introduce the article.
	Hook template.HTML `toml:"hook"`

	// HookImageURL is the URL for a hook image for the article (to be shown on
	// the article index) if one was found.
	HookImageURL string `toml:"-"`

	// Image is an optional image that may be included with an article.
	Image string `toml:"image,omitempty"`

	// Location is the geographical location where this article was written.
	Location string `toml:"location,omitempty" validate:"required"`

	// PublishedAt is when the article was published.
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a unique identifier for the article that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Tag is used to group articles together :)
	Tags []string `toml:"tags,omitempty"`

	// Title is the article's title.
	Title string `toml:"title" validate:"required"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as TOML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC template.HTML `toml:"-"`
}

// publishingInfo produces a brief spiel about publication which is intended to
// go into the left sidebar when an article is shown.
func (a *Article) publishingInfo() map[string]string {
	info := make(map[string]string)

	info["Article"] = a.Title
	info["Published"] = a.PublishedAt.In(localLocation).Format("January 2, 2006")
	info["Location"] = a.Location

	return info
}

func (a *Article) validate(source string) error {
	if err := validate.Struct(a); err != nil {
		return xerrors.Errorf("error validating article %q: %+v", source, err)
	}
	return nil
}

// Page is the metadata for a static HTML page generated from an ACE file.
type Page struct {
	// Paths for external dependencies that the page included as it was being
	// rendered, and which should be watched so that we can re-render it when
	// one changes.
	//
	// Set the first time a page is rendered and updated every subsequent
	// render.
	dependencies []string
}

// articleYear holds a collection of articles grouped by year.
type articleYear struct {
	Year     int
	Articles []*Article
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

// Very similar to RFC 4648 base32 except that numbers come first instead of
// last so that sortable values encoded to base32 will sort in the same
// lexicographic (alphabetical) order as the original values. Also, use lower
// case characters instead of upper.
var lexicographicBase32 = "234567abcdefghijklmnopqrstuvwxyz"

func extCanonical(originalURL string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		panic(err)
	}

	return strings.ToLower(filepath.Ext(u.Path))
}

// Returns a target extension and format given an input one. Currently only used
// to make HEICs (which aren't web friendly) into more widely supported WebPs,
// but I should experiment with more broad use of WebPs. Other formats like JPGs
// and PNGs get left with their input extension/format.
func extImageTarget(canonicalExt string) string {
	if canonicalExt == ".heic" {
		return ".webp"
	}

	return canonicalExt
}

// Gets a map of local values for use while rendering a template and includes
// a few "special" values that are globally relevant to all templates.
func getLocals(locals map[string]interface{}) map[string]interface{} {
	defaults := map[string]interface{}{
		"AbsoluteURL": conf.AbsoluteURL,
		"Release":     Release,
		"SorgEnv":     conf.SorgEnv,
		"TitleSuffix": scommon.TitleSuffix,
	}

	for k, v := range locals {
		defaults[k] = v
	}

	return defaults
}

func groupArticlesByYear(articles []*Article) []*articleYear {
	var year *articleYear
	var years []*articleYear

	for _, article := range articles {
		if year == nil || year.Year != article.PublishedAt.Year() {
			year = &articleYear{article.PublishedAt.Year(), nil}
			years = append(years, year)
		}

		year.Articles = append(year.Articles, article)
	}

	return years
}

func insertOrReplaceArticle(articles *[]*Article, article *Article) {
	for i, a := range *articles {
		if article.Slug == a.Slug {
			(*articles)[i] = article
			return
		}
	}

	*articles = append(*articles, article)
}

func mustLocation(locationName string) *time.Location {
	location, err := time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	return location
}

// Remove the "./pages" directory and extension, but keep the rest of the
// path.
//
// Looks something like "about", or "nested/about".
func pagePathKey(source string) string {
	pagePath := mfile.MustAbs(source)
	pagePath = strings.TrimPrefix(pagePath, mfile.MustAbs("./html/pages")+"/")
	pagePath = strings.TrimSuffix(pagePath, path.Ext(pagePath))
	pagePath = strings.TrimSuffix(pagePath, path.Ext(pagePath)) // again, for `.tmpl.html`
	return pagePath
}

// Checks if the path exists as a common image format (.jpg or .png only). If
// so, returns the discovered extension (e.g. "jpg") and boolean true.
// Otherwise returns an empty string and boolean false.
func pathAsImage(extensionlessPath string) (string, bool) {
	// extensions must be lowercased
	formats := []string{"jpg", "png"}

	for _, format := range formats {
		_, err := os.Stat(extensionlessPath + "." + format)
		if err != nil {
			continue
		}

		return format, true
	}

	return "", false
}

func renderArticle(ctx context.Context, c *modulir.Context, source string,
	articles *[]*Article, articlesChanged *bool, mu *sync.Mutex,
) (bool, error) {
	sourceChanged := c.Changed(source)

	sourceTmpl := scommon.HTML + "/article.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !sourceChanged && !htmlChanged {
		return false, nil
	}

	var article Article
	data, err := mtoml.ParseFileFrontmatter(c, source, &article)
	if err != nil {
		return true, err
	}

	err = article.validate(source)
	if err != nil {
		return true, err
	}

	article.Slug = scommon.ExtractSlug(source)

	content, err := mmarkdownext.Render(string(data), &mmarkdownext.RenderOptions{
		TemplateData: map[string]interface{}{
			"Ctx": ctx,
		},
	})
	if err != nil {
		return true, err
	}

	content, footnotes, ok := strings.Cut(content, `<div class="footnotes">`)
	if ok {
		footnotes = strings.TrimSuffix(footnotes, "</div>")
	}

	article.Content = template.HTML(content)
	article.Footnotes = template.HTML(footnotes) // may be empty

	toc, err := mtoc.RenderFromHTML(string(article.Content))
	if err != nil {
		return true, err
	}

	article.TOC = template.HTML(toc)

	if article.Hook != "" {
		hook, err := mmarkdownext.Render(string(article.Hook), nil)
		if err != nil {
			return true, err
		}

		article.Hook = template.HTML(mtemplate.CollapseParagraphs(hook))
	}

	format, ok := pathAsImage(
		path.Join(c.SourceDir, "content", "images", article.Slug, "hook"),
	)
	if ok {
		article.HookImageURL = "/content/images/" + article.Slug + "/hook." + format
	}

	locals := getLocals(map[string]interface{}{
		"Article":        article,
		"PublishingInfo": article.publishingInfo(),
	})

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, article.Slug), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplaceArticle(articles, &article)
	*articlesChanged = true
	mu.Unlock()

	return true, nil
}

var markdownLinkRE = regexp.MustCompile(`\[(.*?)\]\(.*?\)`)

func simplifyMarkdownForSummary(str string) string {
	str = markdownLinkRE.ReplaceAllString(str, "$1")
	str = strings.ReplaceAll(str, "\n\n", " ")
	str = strings.ReplaceAll(str, "\n", " ")
	return strings.TrimSpace(str)
}

func truncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[0:maxLength-2] + " …"
}

func renderHome(ctx context.Context, c *modulir.Context,
	articles []*Article,
	articlesChanged bool,
	tagMap map[string][]*Article,
) (bool, error) {
	sourceTmpl := scommon.HTML + "/index.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !htmlChanged {
		return false, nil
	}

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals(map[string]interface{}{
		"ArticlesByYear": articlesByYear,
		"TagMap":         tagMap,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "index.html"), locals)
}

func renderTag(ctx context.Context, c *modulir.Context,
	tag string,
	articles []*Article,
	articlesChanged bool,
) (bool, error) {
	sourceTmpl := scommon.HTML + "/tags/tag.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !htmlChanged {
		return false, nil
	}

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals(map[string]interface{}{
		"Tag":            tag,
		"ArticlesByYear": articlesByYear,
	})

	targetDir := path.Join(c.TargetDir, "tags")

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(targetDir, tagToURL(tag)), locals)
}

func renderPage(ctx context.Context, c *modulir.Context,
	source string, meta map[string]*Page, mu *sync.RWMutex,
) (bool, error) {
	pagePath := pagePathKey(source)

	// Other dependencies a page might have if it say, included an external
	// Markdown file. These are added the first time a page is rendered (and
	// watched), and updated on every subsequent run.
	var pageDependencies []string

	mu.RLock()
	pageMeta, ok := meta[pagePath]
	if ok {
		pageDependencies = pageMeta.dependencies
	}
	mu.RUnlock()

	htmlChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			source,
		},
		append(
			universalSources,
			pageDependencies...,
		)...,
	)...)
	if !htmlChanged {
		return false, nil
	}

	// Looks something like "./public/about".
	target := path.Join(c.TargetDir, pagePath)

	// Put a ".html" on if this page is an index. This will allow our local
	// server to serve it at a directory path, and our upload script is smart
	// enough to do the right thing with it as well.
	if path.Base(pagePath) == "index" {
		target += ".html"
	}

	// Reuse existing metadata for this page, or create metadata if this is the
	// first time we're rendering it.
	if pageMeta == nil {
		pageMeta = &Page{}

		mu.Lock()
		meta[pagePath] = pageMeta
		mu.Unlock()
	}

	// Pages get their titles by using inner templates. That must be triggered
	// by sending an empty string as `Title`.
	err := mfile.EnsureDir(c, path.Dir(target))
	if err != nil {
		return true, err
	}

	pageMeta.dependencies = nil

	locals := getLocals(nil)

	err = dependencies.renderGoTemplate(ctx, c, source, target, locals)
	if err != nil {
		return true, err
	}

	pageMeta.dependencies = dependencies.getDependencies(source)

	return true, nil
}

func getTagMap(articles []*Article) map[string][]*Article {
	tagMap := make(map[string][]*Article)
	for _, article := range articles {
		for _, tag := range article.Tags {
			tagMap[tag] = append(tagMap[tag], article)
		}
	}
	return tagMap
}

func tagToURL(tag string) string {
	// Convert to lowercase
	tag = strings.ToLower(tag)

	// Remove all non-word characters and replace with a dash
	re := regexp.MustCompile(`[\s\W-]+`)
	tag = re.ReplaceAllString(tag, "-")

	// Trim leading and trailing dashes
	tag = strings.Trim(tag, "-")

	return tag
}
