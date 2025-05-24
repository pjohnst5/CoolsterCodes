package main

import (
	"context"
	"fmt"
	"html/template"
	"math/rand/v2"
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

	"github.com/brandur/modulir"
	"github.com/brandur/modulir/modules/matom"
	"github.com/brandur/modulir/modules/mfile"
	"github.com/brandur/modulir/modules/mimage"
	"github.com/brandur/modulir/modules/mmarkdownext"
	"github.com/brandur/modulir/modules/mtemplate"
	"github.com/brandur/modulir/modules/mtoc"
	"github.com/brandur/modulir/modules/mtoml"
	"github.com/brandur/sorg/modules/scommon"
	"github.com/brandur/sorg/modules/squantified"
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

// A set of tag constants to hopefully help ensure that this set doesn't grow
// very much.
const (
	tagPostgres Tag = "postgres"
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
	dependencies = NewDependencyRegistry()
	pages        = make(map[string]*Page)
	photos       []*Photo
	photosOther  []*Photo
)

// Time zone to show articles / fragments / etc. publishing times in.
var localLocation = mustLocation("America/Denver")

// List of common build dependencies, a change in any of which will trigger a
// rebuild on everything: partial views, JavaScripts, and stylesheets. Even
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

//nolint:maintidx
func build(c *modulir.Context) []error {
	//
	// PHASE 0: Setup
	//
	// (No jobs should be enqueued here.)
	//

	c.Log.Debugf("Running build loop")

	// This is where we stored "versioned" assets like compiled JS and CSS.
	// These assets have a release number that we can increment and by
	// extension quickly invalidate.
	versionedAssetsDir := path.Join(c.TargetDir, "assets", Release)

	// A set of source paths that rebuild everything when any one of them
	// changes. These are dependencies that are included in more or less
	// everything: common partial views, JavaScript sources, and stylesheet
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

	// Generate a list of partial views to add to universal sources.
	{
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/views",
			&mfile.ReadDirOptions{ShowMeta: true})
		if err != nil {
			return []error{err}
		}

		var partialViews []string
		for _, source := range sources {
			if strings.HasPrefix(filepath.Base(source), "_") {
				partialViews = append(partialViews, source)
			}
		}

		universalSources = append(universalSources, partialViews...)
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

	ctx, downloadedImageContainer := mtemplate.DownloadedImageContext(ctx)

	//
	// Common directories
	//
	// Create these outside of the job system because jobs below may depend on
	// their existence.
	//

	{
		commonDirs := []string{
			c.TargetDir + "/articles",
			c.TargetDir + "/photos",
			c.TargetDir + "/reading",
			scommon.TempDir,
			versionedAssetsDir,
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
			{c.SourceDir + "/content/fonts", c.TargetDir + "/assets/fonts"},
			{c.SourceDir + "/content/images", c.TargetDir + "/assets/images"},
			{c.SourceDir + "/content/javascripts", versionedAssetsDir + "/javascripts"},
			{c.SourceDir + "/content/photographs", c.TargetDir + "/photographs"},
			{c.SourceDir + "/content/stylesheets", versionedAssetsDir + "/stylesheets"},
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
		sources, err := mfile.ReadDirCached(c, c.SourceDir+"/pages", &mfile.ReadDirOptions{RecurseDirs: true})
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
	// Photos (read `_meta.toml`)
	//

	var photosChanged bool

	{
		c.AddJob("photos _meta.toml", func() (bool, error) {
			source := c.SourceDir + "/content/photographs/_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			var photosWrapper PhotoWrapper
			err := mtoml.ParseFile(c, source, &photosWrapper)
			if err != nil {
				return true, err
			}

			photos = photosWrapper.Photos
			photosChanged = true
			return true, nil
		})
	}

	//
	// Photos (other) (read `_other_meta.toml`)
	//

	{
		c.AddJob("photos (other) _meta.toml", func() (bool, error) {
			source := c.SourceDir + "/content/photographs/_other_meta.toml"

			if !c.Changed(source) {
				return false, nil
			}

			var photosWrapper PhotoWrapper
			err := mtoml.ParseFile(c, source, &photosWrapper)
			if err != nil {
				return true, err
			}

			if err := photosWrapper.validate(); err != nil {
				return true, err
			}

			photosOther = photosWrapper.Photos
			return true, nil
		})
	}

	//
	// Reading
	//

	{
		c.AddJob("reading", func() (bool, error) {
			return renderReading(ctx, c)
		})
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
		slices.SortFunc(photos, func(a, b *Photo) int { return b.OccurredAt.Compare(a.OccurredAt) })
	}

	//
	// Articles
	//

	// Index
	{
		c.AddJob("articles index", func() (bool, error) {
			return renderArticlesIndex(ctx, c, articles,
				articlesChanged)
		})
	}

	// Feed (all)
	{
		c.AddJob("articles feed", func() (bool, error) {
			return renderArticlesFeed(c, articles, nil,
				articlesChanged)
		})
	}

	// Feed (Postgres)
	{
		c.AddJob("articles feed (postgres)", func() (bool, error) {
			return renderArticlesFeed(c, articles, tagPointer(tagPostgres),
				articlesChanged)
		})
	}

	//
	// Home
	//

	{
		c.AddJob("home", func() (bool, error) {
			return renderHome(ctx, c, articles, photos,
				articlesChanged, photosChanged)
		})
	}

	//
	// Photos (index / fetch + resize)
	//

	// Photo index
	{
		c.AddJob("photos index", func() (bool, error) {
			return renderPhotoIndex(ctx, c, photos,
				photosChanged)
		})
	}

	// Photo fetch + resize
	{
		for _, p := range photos {
			photo := p

			name := "photo: " + photo.Slug
			c.AddJob(name, func() (bool, error) {
				return fetchAndResizePhoto(c, c.SourceDir+"/content/photographs", photo)
			})
		}
	}

	// Photo fetch + resize (other)
	{
		for _, p := range photosOther {
			photo := p

			name := "photo fetch: " + photo.Slug
			c.AddJob(name, func() (bool, error) {
				return fetchAndResizePhotoOther(c, c.SourceDir+"/content/photographs", photo)
			})
		}
	}

	// From `DownloadedImage` template tags.
	{
		for i := range downloadedImageContainer.Images {
			imageInfo := downloadedImageContainer.Images[i]

			c.AddJob("downloaded image: "+imageInfo.Slug, func() (bool, error) {
				return fetchAndResizeDownloadedImage(c, c.SourceDir+"/content/photographs", imageInfo)
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

	// Tags are the set of tags that the article is tagged with.
	Tags []Tag `toml:"tags,omitempty"`

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

// taggedWith returns true if the given tag is in this article's set of tags
// and false otherwise.
func (a *Article) taggedWith(tag Tag) bool {
	for _, t := range a.Tags {
		if t == tag {
			return true
		}
	}

	return false
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

// Photo is a photograph.
type Photo struct {
	// CropGravity is the gravity to use with ImageMagick when doing a square
	// crop. Should be one of: northwest, north, northeast, west, center, east,
	// southwest, south, southeast.
	CropGravity string `default:"center" toml:"crop_gravity"`

	// CropWidth is the width to crop the photo to.
	//
	// This should be the non-retina target width. A second file will be
	// created with the `@2x` suffix with twice this number.
	//
	// This is a required property for photos that are not part of the main
	// photographs sequence. It's ignored for photos that *are* part of the
	// main photographs sequence.
	CropWidth int `toml:"crop_width"`

	// Description is the description of the photograph.
	Description string `toml:"description"`

	// KeepInHomeRotation is a special override for photos I really like that
	// keeps them in the home page's random rotation. The rotation then
	// consists of either a recent photo or one of these explicitly selected
	// old ones.
	KeepInHomeRotation bool `toml:"keep_in_home_rotation"`

	// LinkURL is a URL to have the image link to. This is only respect for some
	// uses of photographs like in atoms.
	LinkURL string `toml:"link_url" validate:"-"`

	// NoCrop disables cropping on this photo (normally photos are cropped to
	// 3:2 or 2:3).
	NoCrop bool `toml:"no_crop"`

	// OriginalImageURL is the location where the original-sized version of the
	// photo can be downloaded from.
	OriginalImageURL string `toml:"original_image_url" validate:"required"`

	// OccurredAt is UTC time when the photo was published.
	OccurredAt time.Time `toml:"occurred_at"`

	// OverrideExt is an extension like `.webp` that should be used for the
	// resized versions of the photo. Mostly useful for when a screenshot or
	// something is saved as a `.png` and it should really have been a `.jpg` or
	// something because the source being displayed was already lossy.
	OverrideExt string `toml:"override_ext" validate:"-"`

	// Portrait is a hint to indicate that the photo is in portrait instead of
	// landscape. This helps the build pick a better stand-in image for lazy
	// loading so that there's less jumping around as photos that get loaded in
	// change size.
	Portrait bool `toml:"portrait"`

	// Slug is a unique identifier for the photo. Originally these were
	// generated from Flickr, but I've since just started reusing them for
	// filenames.
	Slug string `toml:"slug" validate:"required"`

	// Title is the title of the photograph.
	Title string `toml:"title"`

	// Internal
	originalExt string `toml:"-"`
}

func (p *Photo) Equal(other *Photo) bool {
	return p.CropGravity == other.CropGravity &&
		p.CropWidth == other.CropWidth &&
		p.Description == other.Description &&
		p.KeepInHomeRotation == other.KeepInHomeRotation &&
		p.LinkURL == other.LinkURL &&
		p.NoCrop == other.NoCrop &&
		p.OriginalImageURL == other.OriginalImageURL &&
		p.OccurredAt.Equal(other.OccurredAt) &&
		p.OverrideExt == other.OverrideExt &&
		p.Portrait == other.Portrait &&
		p.Slug == other.Slug &&
		p.Title == other.Title
}

func (p *Photo) OriginalExt() string {
	if p.originalExt != "" {
		return p.originalExt
	}

	p.originalExt = extCanonical(p.OriginalImageURL)
	return p.originalExt
}

func (p *Photo) TargetExt() string {
	if p.OverrideExt != "" {
		return p.OverrideExt
	}

	return extImageTarget(p.OriginalExt())
}

// PhotoWrapper is a data structure intended to represent the data structure at
// the top level of photograph data file `content/photographs/_meta.toml`.
type PhotoWrapper struct {
	// Photos is a collection of photos within the top-level wrapper.
	Photos []*Photo `toml:"photographs" validate:"required,dive"`
}

func (w *PhotoWrapper) validate() error {
	if err := validate.Struct(w); err != nil {
		return xerrors.Errorf("error validating photos: %+v", err)
	}
	return nil
}

// Tag is a symbol assigned to an article to categorize it.
//
// This feature is not meanted to be overused. It's really just for tagging
// a few particular things so that we can generate content-specific feeds for
// certain aggregates (so far just Planet Postgres).
type Tag string

// articleYear holds a collection of articles grouped by year.
type articleYear struct {
	Year     int
	Articles []*Article
}

// readingYear holds a collection of readings grouped by year.
type readingYear struct {
	Year     int
	Readings []*squantified.Reading
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

var cropDefault = &mimage.PhotoCropSettings{Portrait: "2:3", Landscape: "3:2"}

var defaultPhotoSizes = []mimage.PhotoSize{
	{Suffix: "", Width: 333, CropSettings: cropDefault},
	{Suffix: "@2x", Width: 667, CropSettings: cropDefault},
	{Suffix: "_large", Width: 1500, CropSettings: cropDefault},
	{Suffix: "_large@2x", Width: 3000, CropSettings: cropDefault},
}

var defaultPhotoSizesNoCrop = []mimage.PhotoSize{
	{Suffix: "", Width: 333, CropSettings: nil},
	{Suffix: "@2x", Width: 667, CropSettings: nil},
	{Suffix: "_large", Width: 1500, CropSettings: nil},
	{Suffix: "_large@2x", Width: 3000, CropSettings: nil},
}

func fetchAndResizePhoto(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for photo '%s': %w", photo.Slug, err)
	}

	photoSizes := defaultPhotoSizes
	if photo.NoCrop {
		photoSizes = defaultPhotoSizesNoCrop
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug, photo.TargetExt(),
		mimage.PhotoGravity(photo.CropGravity), photoSizes)
}

func fetchAndResizeDownloadedImage(c *modulir.Context,
	targetDir string, imageInfo *mtemplate.DownloadedImageInfo,
) (bool, error) {
	base := filepath.Base(imageInfo.Slug)
	dir := targetDir + filepath.Dir(imageInfo.Slug)

	extImageTarget := func(canonicalExt string) string {
		if canonicalExt == ".heic" {
			return ".webp"
		}

		return canonicalExt
	}

	return mimage.FetchAndResizeImage(c, imageInfo.URL, dir, base, extImageTarget(imageInfo.OriginalExt()), mimage.PhotoGravityCenter,
		[]mimage.PhotoSize{
			{Suffix: "", Width: imageInfo.Width, CropSettings: cropDefault},
			{Suffix: "@2x", Width: imageInfo.Width * 2, CropSettings: cropDefault},
		})
}

func fetchAndResizePhotoOther(c *modulir.Context, targetDir string, photo *Photo) (bool, error) {
	if photo.CropWidth == 0 {
		return false, xerrors.Errorf("need `crop_width` specified for photo '%s'", photo.Slug)
	}

	u, err := url.Parse(photo.OriginalImageURL)
	if err != nil {
		return false, xerrors.Errorf("bad URL for photo '%s'", photo.Slug)
	}

	return mimage.FetchAndResizeImage(c, u, targetDir, photo.Slug, photo.TargetExt(),
		mimage.PhotoGravity(photo.CropGravity),
		[]mimage.PhotoSize{
			{Suffix: "", Width: photo.CropWidth, CropSettings: nil},
			{Suffix: "@2x", Width: photo.CropWidth * 2, CropSettings: nil},
		})
}

// Gets a map of local values for use while rendering a template and includes
// a few "special" values that are globally relevant to all templates.
func getLocals(locals map[string]interface{}) map[string]interface{} {
	defaults := map[string]interface{}{
		"AbsoluteURL":       conf.AbsoluteURL,
		"EnableGoatCounter": conf.EnableGoatCounter,
		"GoogleAnalyticsID": conf.GoogleAnalyticsID,
		"LocalFonts":        conf.LocalFonts,
		"Release":           Release,
		"SorgEnv":           conf.SorgEnv,
		"TitleSuffix":       scommon.TitleSuffix,
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

func groupReadingsByYear(readings []*squantified.Reading) []*readingYear {
	var year *readingYear
	var years []*readingYear

	for _, reading := range readings {
		if year == nil || year.Year != reading.ReadAt.Year() {
			year = &readingYear{reading.ReadAt.Year(), nil}
			years = append(years, year)
		}

		year.Readings = append(year.Readings, reading)
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
	pagePath = strings.TrimPrefix(pagePath, mfile.MustAbs("./pages")+"/")
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

	sourceTmpl := scommon.ViewsDir + "/articles/show.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !sourceChanged && !viewsChanged {
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
		article.HookImageURL = "/assets/images/" + article.Slug + "/hook." + format
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

func renderArticlesIndex(ctx context.Context, c *modulir.Context, articles []*Article, articlesChanged bool) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/articles/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !viewsChanged {
		return false, nil
	}

	articlesByYear := groupArticlesByYear(articles)

	locals := getLocals(map[string]interface{}{
		"ArticlesByYear": articlesByYear,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "articles/index.html"), locals)
}

func renderArticlesFeed(_ *modulir.Context, articles []*Article, tag *Tag, articlesChanged bool) (bool, error) {
	if !articlesChanged {
		return false, nil
	}

	name := "articles"
	if tag != nil {
		name = fmt.Sprintf("articles-%s", *tag)
	}
	atomPath := name + ".atom"

	title := "Articles" + scommon.TitleSuffix
	if tag != nil {
		title = fmt.Sprintf("Articles%s (%s)", scommon.TitleSuffix, *tag)
	}

	feed := &matom.Feed{
		Title: title,
		ID:    "tag:" + scommon.AtomTag + ",2013:/" + name,

		Links: []*matom.Link{
			{Rel: "self", Type: "application/atom+xml", Href: "https://brandur.org/" + atomPath},
			{Rel: "alternate", Type: "text/html", Href: "https://brandur.org"},
		},
	}

	if len(articles) > 0 {
		feed.Updated = articles[0].PublishedAt
	}

	for i, article := range articles {
		if tag != nil && !article.taggedWith(*tag) {
			continue
		}

		if i >= conf.NumAtomEntries {
			break
		}

		entry := &matom.Entry{
			Title:     article.Title,
			Summary:   string(article.Hook),
			Content:   &matom.EntryContent{Content: string(article.Content), Type: "html"},
			Published: article.PublishedAt,
			Updated:   article.PublishedAt,
			Link:      &matom.Link{Href: conf.AbsoluteURL + "/" + article.Slug},
			ID:        "tag:" + scommon.AtomTag + "," + article.PublishedAt.Format("2006-01-02") + ":" + article.Slug,

			AuthorName: scommon.AtomAuthorName,
			AuthorURI:  conf.AbsoluteURL,
		}
		feed.Entries = append(feed.Entries, entry)
	}

	filename := path.Join(conf.TargetDir, atomPath)
	f, err := os.Create(filename)
	if err != nil {
		return true, xerrors.Errorf("error creating file '%s': %w", filename, err)
	}
	defer f.Close()

	return true, feed.Encode(f, "  ")
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
	articles []*Article, photos []*Photo,
	articlesChanged, photosChanged bool,
) (bool, error) {
	sourceTmpl := scommon.ViewsDir + "/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(sourceTmpl)...)
	if !articlesChanged && !photosChanged && !viewsChanged {
		return false, nil
	}

	if len(articles) > 3 {
		articles = articles[0:3]
	}

	// Find a random photo to put on the homepage.
	photo := selectRandomPhoto(photos)

	locals := getLocals(map[string]interface{}{
		"Articles": articles,
		"Photo":    photo,
	})

	return true, dependencies.renderGoTemplate(ctx, c, sourceTmpl,
		path.Join(c.TargetDir, "index.html"), locals)
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

	viewsChanged := c.ChangedAny(append(
		[]string{
			scommon.MainLayout,
			source,
		},
		append(
			universalSources,
			pageDependencies...,
		)...,
	)...)
	if !viewsChanged {
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

func renderReading(ctx context.Context, c *modulir.Context) (bool, error) {
	source := scommon.ViewsDir + "/reading/index.tmpl.html"
	viewsChanged := c.ChangedAny(
		append([]string{
			c.SourceDir + "/content/reading/_meta.toml",
		},
			dependencies.getDependencies(source)...,
		)...)
	if !c.FirstRun && !viewsChanged {
		return false, nil
	}

	readings, err := squantified.GetReadingsData(c, c.SourceDir+"/content/reading/_meta.toml")
	if err != nil {
		return false, err
	}

	readingsByYear := groupReadingsByYear(readings)

	locals := getLocals(map[string]interface{}{
		"ReadingsByYear": readingsByYear,
	})

	err = dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "reading/index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func renderPhotoIndex(ctx context.Context, c *modulir.Context, photos []*Photo,
	photosChanged bool,
) (bool, error) {
	source := scommon.ViewsDir + "/photos/index.tmpl.html"
	viewsChanged := c.ChangedAny(dependencies.getDependencies(source)...)
	if !photosChanged && !viewsChanged {
		return false, nil
	}

	locals := getLocals(map[string]interface{}{
		"Photos": photos,
	})

	err := dependencies.renderGoTemplate(ctx, c, source, path.Join(c.TargetDir, "photos", "index.html"), locals)
	if err != nil {
		return true, err
	}

	return true, nil
}

func selectRandomPhoto(photos []*Photo) *Photo {
	if len(photos) < 1 {
		return nil
	}

	numRecent := 20
	if len(photos) < numRecent {
		numRecent = len(photos)
	}

	// All recent photos go into the random selection.
	randomPhotos := photos[0:numRecent]

	// Older photos that are good enough that I've explicitly tagged them
	// as such also get considered for the rotation.
	if len(photos) > numRecent {
		olderPhotos := photos[numRecent : len(photos)-1]

		for _, photo := range olderPhotos {
			if photo.KeepInHomeRotation {
				randomPhotos = append(randomPhotos, photo)
			}
		}
	}

	//nolint:gosec
	return randomPhotos[rand.IntN(len(randomPhotos))]
}

// Gets a pointer to a tag just to work around the fact that you can take the
// address of a constant like `tagPostgres`.
func tagPointer(tag Tag) *Tag {
	return &tag
}
