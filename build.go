package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	stripmd "github.com/writeas/go-strip-markdown"
	"golang.org/x/xerrors"

	"coolstercodes/modules/modulir"
	"coolstercodes/modules/modulir/mfile"
	"coolstercodes/modules/modulir/mmarkdownext"
	"coolstercodes/modules/modulir/mtemplate"
	"coolstercodes/modules/modulir/mtoc"
	"coolstercodes/modules/modulir/mtoml"
	"coolstercodes/modules/scommon"
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

const (
	NTags = 10
	MTags = 2
)

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
	pages        []*Page
	dependencies = NewDependencyRegistry()
)

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

//nolint:gocyclo,maintidx // complexity is acceptable for this case
func build(c *modulir.Context) []error {
	//
	// PHASE 0: Setup
	//
	// (No jobs should be enqueued here.)
	//

	c.Log.Debugf("Running build loop")

	// This is where we stored content like compiled JS and CSS.
	contentDir := path.Join(c.TargetDir, "content")

	// A set of source paths that rebuild everything when any one of them
	// changes. These are dependencies that are included in more or less
	// everything: common partial html, JavaScript sources, and stylesheet
	// sources.
	universalSources = nil

	// Generate a set of JavaScript sources to add to universal sources.
	{
		javaScriptSources, err := mfile.ReadDirCached(c, c.SourceDir+"/web/javascripts",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}
		universalSources = append(universalSources, javaScriptSources...)
	}

	// Generate a list of partial html to add to universal sources.
	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/web/html",
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
		stylesheetSources, err := mfile.ReadDirCached(c, c.SourceDir+"/web/stylesheets",
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
			contentDir,
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
			{c.SourceDir + "/web/javascripts", contentDir + "/javascripts"},
			{c.SourceDir + "/web/stylesheets", contentDir + "/stylesheets"},
		}
		for _, link := range commonSymlinks {
			err := mfile.EnsureSymlink(c, link[0], link[1])
			if err != nil {
				return []error{nil}
			}
		}
	}

	//
	// Recursively copy over article pictures into /content/images
	//

	if err := mfile.CopyDirectoryImages(c, c.SourceDir+"/content/articles", c.TargetDir+"/content/images"); err != nil {
		return []error{err}
	}

	//
	// Articles
	//

	var articlesChanged bool
	var articlesMu sync.Mutex

	{
		opts := mfile.ReadDirOptions{
			RecurseDirs: true,
			OnlyGetMDs:  true,
		}
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/articles", &opts)
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
	// Recursively copy over pages pictures into /content/images
	//

	if err := mfile.CopyDirectoryImages(c, c.SourceDir+"/content/pages", c.TargetDir+"/content/images"); err != nil {
		return []error{err}
	}

	//
	// Pages (render each view)
	//

	var pagesChanged bool
	var pagesMu sync.RWMutex

	{
		opts := mfile.ReadDirOptions{
			RecurseDirs: true,
			OnlyGetMDs:  true,
		}
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/content/pages", &opts)
		if err != nil {
			return []error{err}
		}

		for _, s := range sources {
			source := s

			name := "page: " + filepath.Base(source)
			c.AddJob(name, func() (bool, error) {
				return renderPage(ctx, c, source,
					&pages, &pagesChanged, &pagesMu)
			})
		}
	}

	//
	// Copy over remaining images to /content/images
	//
	if err := mfile.CopyFile(c, c.SourceDir+"/content/images/CoolsterCodes.png", c.TargetDir+"/content/images/CoolsterCodes.png"); err != nil {
		return []error{err}
	}
	if err := mfile.CopyFile(c, c.SourceDir+"/content/images/favicon.png", c.TargetDir+"/content/images/favicon.png"); err != nil {
		return []error{err}
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

	tagMap := getTagMap(articles)
	tagCount := getAllTagCounts(tagMap)
	topNTags, topMTags := getTopNAndMTags(tagCount, NTags, MTags)
	{
		c.AddJob("home", func() (bool, error) {
			return renderHome(ctx, c, articles,
				articlesChanged, topNTags, topMTags)
		})
	}

	//
	// Tags
	//
	{
		for tag, articles := range tagMap {
			c.AddJob(tag, func() (bool, error) {
				return renderTag(ctx, c,
					tag,
					articles,
					articlesChanged)
			})
		}
	}

	{
		c.AddJob("tags", func() (bool, error) {
			return renderAllTags(ctx, c, tagCount, articlesChanged)
		})
	}

	//
	// Index
	//
	{
		indexFileName := "index.json"
		srcPath := "./web/" + indexFileName
		dstPath := contentDir + "/" + indexFileName
		c.AddJob("index", func() (bool, error) {
			return generateIndex(srcPath, dstPath, articles, pages)
		})
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

	// This would be '/content/images/<slug>/
	ImgDir string `toml:"-"`

	// Footnotes are HTML footnotes extracted from content.
	Footnotes template.HTML `toml:"-"`

	// Hook is a leading sentence or two to succinctly introduce the article.
	Hook template.HTML `toml:"hook"`

	// Image is an optional image that may be included with an article.
	Image string `toml:"image,omitempty"`

	// PublishedAt is when the article was published.
	PublishedAt time.Time `toml:"published_at" validate:"required"`

	// Slug is a unique identifier for the article that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Youtube video link if applicable
	YouTube string `toml:"youtube"`

	// YoutubeEmbed link if applicable
	YouTubeEmbed string `toml:"-"`

	// Tag is used to group articles together :)
	Tags []string `toml:"tags,omitempty"`

	// Both tags as string and url (not really using count)
	TagCounts []TagCount `toml:"tagcounts,omitempty"`

	// Title is the article's title.
	Title string `toml:"title" validate:"required"`

	// TOC is the HTML rendered table of contents of the article. It isn't
	// included as TOML frontmatter, but rather calculated from the article's
	// content, rendered, and then added separately.
	TOC template.HTML `toml:"-"`

	// The searchable body for index.json
	Body string `toml:"body"`
}

type Page struct {
	// Content is the HTML content of the article. It isn't included as TOML
	// frontmatter, and is rather split out of an article's Markdown file,
	// rendered, and then added separately.
	Content template.HTML `toml:"-"`

	// Slug is a unique identifier for the page that also helps determine
	// where it's addressable by URL.
	Slug string `toml:"-"`

	// Title is the article's title.
	Title string `toml:"title" validate:"required"`

	// Description is the metadata description
	Description string `toml:"description" validate:"required"`

	// The searchable body for index.json
	Body string `toml:"body"`

	// This would be '/content/images/<slug>/
	ImgDir string `toml:"-"`
}

type IndexEntry struct {
	Href    string   `json:"href"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Tags    []string `json:"tags"`
	Img     string   `json:"img"`
}

func (a *Article) validate(source string) error {
	if err := validate.Struct(a); err != nil {
		return xerrors.Errorf("error validating article %q: %+v", source, err)
	}
	return nil
}

func (p *Page) validate(source string) error {
	if err := validate.Struct(p); err != nil {
		return xerrors.Errorf("error validating page %q: %+v", source, err)
	}
	return nil
}

type TagCount struct {
	Tag    string
	Count  int
	URLTag string
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
		"FavIcon":     "/content/images/favicon.png",
		"CCEnv":       conf.CCEnv,
		"TitleSuffix": scommon.TitleSuffix,
	}

	for k, v := range locals {
		defaults[k] = v
	}

	return defaults
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

func insertOrReplacePage(pages *[]*Page, page *Page) {
	for i, a := range *pages {
		if page.Slug == a.Slug {
			(*pages)[i] = page
			return
		}
	}
	*pages = append(*pages, page)
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
		return true, xerrors.Errorf("error parsing frontmatter %v", err)
	}

	// Sort tags really quick
	sort.Strings(article.Tags)

	article.Slug = scommon.ExtractSlug(source)
	relativeDir := scommon.GetPathToParentDirectory(source)

	// Define an ImgDir (for later processing) and set Image as full path
	article.ImgDir = "/" + strings.Replace(relativeDir, "articles", "images", 1) + "/"
	if article.Image != "" {
		article.Image = filepath.Join(article.ImgDir, article.Image)
	}
	if article.YouTube != "" {
		article.YouTubeEmbed = getYouTubeEmbedLink(article.YouTube)
	}
	stripped := stripmd.Strip(string(data))
	article.Body = strings.ReplaceAll(stripped, "\n", " ")
	article.Body = strings.ReplaceAll(article.Body, "’", "'")
	article.Body = strings.ReplaceAll(article.Body, "–", "-")
	article.Body = string(article.Hook) + " " + article.Body

	err = article.validate(source)
	if err != nil {
		return true, err
	}

	content, err := mmarkdownext.Render(string(data), &mmarkdownext.RenderOptions{
		TemplateData: map[string]interface{}{
			"Ctx": ctx,
		},
		ImgDir: article.ImgDir,
	})
	if err != nil {
		return true, xerrors.Errorf("error rendering markdown %v", err)
	}

	content, footnotes, ok := strings.Cut(content, `<div class="footnotes">`)
	if ok {
		if i := strings.LastIndex(footnotes, "</div>"); i != -1 {
			footnotes = footnotes[:i]
		}
		footnotes = strings.TrimSpace(footnotes)
	}

	article.Content = template.HTML(content)
	article.Footnotes = template.HTML(footnotes) // may be empty

	toc, err := mtoc.RenderFromHTML(string(article.Content))
	if err != nil {
		return true, xerrors.Errorf("error rendering html %v", err)
	}

	article.TOC = template.HTML(toc)

	if article.Hook != "" {
		hook, err := mmarkdownext.Render(string(article.Hook), nil)
		if err != nil {
			return true, xerrors.Errorf("error rendering hook %v", err)
		}

		article.Hook = template.HTML(mtemplate.CollapseParagraphs(hook))
	}

	for _, tag := range article.Tags {
		article.TagCounts = append(article.TagCounts, TagCount{Tag: tag, URLTag: tagToURL(tag)})
	}

	locals := getLocals(map[string]interface{}{
		"Article": article,
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
	topNTags []TagCount,
	topMTags []TagCount,
) (bool, error) {
	sourceTmpl := scommon.HTML + "/index.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !htmlChanged {
		return false, nil
	}

	locals := getLocals(map[string]interface{}{
		"Articles": articles,
		"TopNTags": topNTags,
		"TopMTags": topMTags,
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

	urlTag := tagToURL(tag)

	locals := getLocals(map[string]interface{}{
		"Tag":      tag,
		"URLTag":   urlTag,
		"Articles": articles,
	})

	targetDir := path.Join(c.TargetDir, "tags")

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(targetDir, tagToURL(tag)), locals)
}

func renderAllTags(ctx context.Context, c *modulir.Context,
	tagCount []TagCount,
	articlesChanged bool,
) (bool, error) {
	srcTmpl := scommon.HTML + "/tags/tags.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(srcTmpl)...)
	if !articlesChanged && !htmlChanged {
		return false, nil
	}

	locals := getLocals(map[string]interface{}{
		"TagCount": tagCount,
	})

	return true, dependencies.renderGoTemplate(ctx, c, srcTmpl,
		path.Join(c.TargetDir, "tags/index.html"), locals)
}

func renderPage(ctx context.Context, c *modulir.Context, source string,
	pages *[]*Page, pagesChanged *bool, mu *sync.RWMutex,
) (bool, error) {
	sourceChanged := c.Changed(source)

	sourceTmpl := scommon.HTML + "/page.tmpl.html"
	htmlChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !sourceChanged && !htmlChanged {
		return false, nil
	}

	var page Page
	data, err := mtoml.ParseFileFrontmatter(c, source, &page)
	if err != nil {
		return true, xerrors.Errorf("error parsing frontmatter %v", err)
	}

	page.Slug = scommon.ExtractSlug(source)
	relativeDir := scommon.GetPathToParentDirectory(source)

	// Define an ImgDir (for later processing) and set Image as full path
	page.ImgDir = "/" + strings.Replace(relativeDir, "pages", "images", 1) + "/"

	stripped := stripmd.Strip(string(data))
	page.Body = strings.ReplaceAll(stripped, "\n", " ")

	err = page.validate(source)
	if err != nil {
		return true, err
	}

	content, err := mmarkdownext.Render(string(data), &mmarkdownext.RenderOptions{
		TemplateData: map[string]interface{}{
			"Ctx": ctx,
		},
		ImgDir: page.ImgDir,
	})
	if err != nil {
		return true, xerrors.Errorf("error rendering markdown %v", err)
	}
	page.Content = template.HTML(content)

	locals := getLocals(map[string]interface{}{
		"Page": page,
	})

	err = dependencies.renderGoTemplate(ctx, c, sourceTmpl, path.Join(c.TargetDir, page.Slug), locals)
	if err != nil {
		return true, err
	}

	mu.Lock()
	insertOrReplacePage(pages, &page)
	*pagesChanged = true
	mu.Unlock()

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

func getTopNAndMTags(tagCount []TagCount, n, m int) ([]TagCount, []TagCount) {
	tagCopy := make([]TagCount, len(tagCount))
	copy(tagCopy, tagCount)

	// Sort by Score (ascending), then Name (alphabetical)
	sort.Slice(tagCopy, func(i, j int) bool {
		if tagCopy[i].Count == tagCopy[j].Count {
			return tagCopy[i].Tag < tagCopy[j].Tag
		}
		return tagCopy[i].Count > tagCopy[j].Count
	})

	if n >= len(tagCopy) {
		return tagCopy, []TagCount{}
	}

	end := n + m
	if end > len(tagCopy) {
		end = len(tagCopy)
	}
	return tagCopy[0:n], tagCopy[n:end]
}

func getAllTagCounts(tagMap map[string][]*Article) []TagCount {
	tagCounts := []TagCount{}
	for tag, articles := range tagMap {
		tagCounts = append(tagCounts, TagCount{
			Tag:    tag,
			Count:  len(articles),
			URLTag: tagToURL(tag),
		})
	}

	// Organizes alphabetically
	sort.Slice(tagCounts, func(i, j int) bool {
		return tagCounts[i].Tag < tagCounts[j].Tag
	})

	return tagCounts
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

func getYouTubeEmbedLink(link string) string {
	// Get everything after the last "/"
	id := link[strings.LastIndex(link, "/")+1:]
	return "https://www.youtube.com/embed/" + id
}

func generateIndex(srcPath, dstPath string, articles []*Article, pages []*Page) (bool, error) {
	entries := map[string]IndexEntry{}
	for _, a := range articles {
		entries[a.Slug] = IndexEntry{
			Href:    a.Slug,
			Title:   a.Title,
			Summary: a.Body,
			Tags:    a.Tags,
			Img:     a.Image,
		}
	}

	for _, p := range pages {
		entries[p.Slug] = IndexEntry{
			Href:    p.Slug,
			Title:   p.Title,
			Summary: p.Body,
		}
	}

	file, err := os.Create(srcPath)
	if err != nil {
		return false, xerrors.Errorf("error creating src file %s: %v", srcPath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ") // pretty-print
	if err := encoder.Encode(entries); err != nil {
		return false, xerrors.Errorf("error encoding %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		return false, xerrors.Errorf("error copying file %v", err)
	}
	return true, nil
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return xerrors.Errorf("error creating src file %s: %v", src, err)
	}
	defer sourceFile.Close() // Ensure the source file is closed

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return xerrors.Errorf("error creating dst file %s: %v", dst, err)
	}
	defer destinationFile.Close() // Ensure the destination file is closed

	// Copy the contents
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return xerrors.Errorf("error copying file %v", err)
	}

	return nil
}
