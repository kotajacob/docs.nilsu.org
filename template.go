package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
	text "text/template"
)

// parseTemplateFiles reads a list of filenames to create a template
func parseTemplateFiles(filenames ...string) *template.Template {
	var files []string
	t := template.New("layout")
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}
	t = template.Must(t.ParseFiles(files...))
	return t
}

// writeHTML to a http.ResponseWriter based on a list of template files
func writeHTML(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
}

// writeText to a http.ResponseWriter based on a list of template files
func writeText(w http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, path.Join(config.TemplateDir, fmt.Sprintf("%s.html", file)))
	}

	templates := text.Must(text.ParseFiles(files...))
	templates.ExecuteTemplate(w, "layout", data)
}
