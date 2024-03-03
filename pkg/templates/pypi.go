package templates

import (
	"html/template"
	"io"
	"reflect"

	"github.com/labstack/echo/v4"
)

const PypiHTML = `<!DOCTYPE html>
<html lang="en">
<head><title>Links for {{ .Name }}</title>
  <meta name="api-version" value="2"/>
</head>
<body><h1>Links for {{ .Name }}</h1>
{{range .Files}}  <a href="{{ .URL }}#sha256={{ .Hashes.Sha256 }}" rel="internal" {{if eq (kindIs .Yanked "bool") false}}data-yanked="{{.Yanked}}"{{end}} {{ $length := len .RequiresPython }} {{ if eq $length 0 }}{{else}}data-requires-python="{{.RequiresPython}}"{{end}}>{{ .Filename }}</a><br/>
{{end}}
</body>
</html>
`

type TemplateRegistry struct {
	Templates *template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.Templates.ExecuteTemplate(w, name, data)
}

func KindIs(v interface{}, kind string) bool {
	return reflect.TypeOf(v).Kind().String() == kind
}
