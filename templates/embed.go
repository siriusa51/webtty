package templates

import (
	"embed"
	"html/template"
)

//go:embed *
var fs embed.FS

// GetTemplate
func GetTemplate(subpath string) *template.Template {
	tmpl, err := template.ParseFS(fs, subpath)
	if err != nil {
		panic(err)
	}

	return tmpl
}

func GetFile(path string) ([]byte, error) {
	return fs.ReadFile(path)
}
