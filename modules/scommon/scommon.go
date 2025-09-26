package scommon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"coolstercodes/modules/modulir/mtemplate"
	"coolstercodes/modules/modulir/mtemplatemd"
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
	// LayoutsDir is the source directory for view layouts.
	LayoutsDir = "./web/html/layouts"

	// MainLayout is the site's main layout in the deprecated ACE templating
	// system. This is no longer used except in a few near retired pages like
	// runs and Twitter.
	MainLayout = LayoutsDir + "/main.ace"

	// TitleSuffix is the suffix to add to the end of page and Atom titles.
	TitleSuffix = " - Coolster Codes"

	// HTML is the source directory for html.
	HTML = "./web/html"
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

func GetPathToParentDirectory(source string) string {
	dir := filepath.Dir(source)
	return dir
}
