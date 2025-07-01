package main

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDependencyRegistryParseGoTemplate(t *testing.T) {
	dependencies := NewDependencyRegistry()
	emptyTmpl := template.New("base_empty")

	// Use some preexistting template for simplicity.
	{
		_, dependencies, err := dependencies.parseGoTemplate(template.Must(emptyTmpl.Clone()), "web/html/layouts/main.tmpl.html")
		require.NoError(t, err)
		require.Equal(t, []string{
			"web/html/layouts/main.tmpl.html",
			"web/html/helpers/_style_stylesheets.tmpl.html",
			"web/html/helpers/_shiki_js.tmpl.html",
			"web/html/helpers/_fancybox_js.tmpl.html",
			"web/html/helpers/_buttons_js.tmpl.html",
			"web/html/helpers/_search_js.tmpl.html",
		}, dependencies)
	}
}

func TestFindGoSubTemplates(t *testing.T) {
	require.Equal(t, []string{"web/html/layouts/main.tmpl.html"}, findGoSubTemplates(`{{template "web/html/layouts/main.tmpl.html" .}}`))
	require.Equal(t, []string{"web/html/layouts/main.tmpl.html"}, findGoSubTemplates(`{{template "web/html/layouts/main.tmpl.html" .}}`))
	require.Equal(t,
		[]string{"web/html/layouts/main.tmpl.html", "web/html/_other.tmpl.html"},
		findGoSubTemplates(`{{template "web/html/layouts/main.tmpl.html" .}}{{template "web/html/_other.tmpl.html" .}}`),
	)
	require.Equal(t, []string{}, findGoSubTemplates(`no templates here`))
	require.Equal(t, []string{}, findGoSubTemplates(``))
}
