package templates

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var content embed.FS

func HTMLRender(w http.ResponseWriter, tmpl string) {
	render, err := template.ParseFS(content, tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html")

	err = render.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
