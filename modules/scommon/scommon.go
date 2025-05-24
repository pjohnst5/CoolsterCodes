package scommon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/brandur/modulir/modules/mtemplate"
	"github.com/brandur/modulir/modules/mtemplatemd"
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

const (
	// AtomAuthorName is the name of the author to include in Atom feeds.
	AtomAuthorName = "Brandur Leach"

	// AtomTag is a stable constant to use in Atom tags.
	AtomTag = "brandur.org"

	// DataDir is where various TOML files for quantified self statistics
	// reside. These are pulled from another project which updates them
	// automatically.
	DataDir = "./data"

	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./layouts"

	// MainLayout is the site's main layout in the deprecated ACE templating
	// system. This is no longer used except in a few near retired pages like
	// runs and Twitter.
	MainLayout = LayoutsDir + "/main.ace"

	// TempDir is a temporary directory used to download images that will be
	// processed and such.
	TempDir = "./tmp"

	// TitleSuffix is the suffix to add to the end of page and Atom titles.
	TitleSuffix = " — brandur.org"

	// ViewsDir is the source directory for views.
	ViewsDir = "./views"
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

// HTMLTemplateFuncMap is a function map of template helpers which is the
// combined version of the maps from ftemplate, mtemplate, and mtemplatemd.
var HTMLTemplateFuncMap = mtemplate.CombineFuncMaps(
	stemplate.FuncMap,
	mtemplate.FuncMap,
	mtemplatemd.FuncMap,
)

// TextTemplateFuncMap is a combined set of template helpers for text
// templates.
var TextTemplateFuncMap = mtemplate.HTMLFuncMapToText(HTMLTemplateFuncMap)

//////////////////////////////////////////////////////////////////////////////
//
//
//
// Functions
//
//
//
//////////////////////////////////////////////////////////////////////////////

// ExitWithError prints the given error to stderr and exits with a status of 1.
func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

// ExtractSlug gets a slug for the given filename by using its basename
// stripped of file extension.
func ExtractSlug(source string) string {
	return strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
}
