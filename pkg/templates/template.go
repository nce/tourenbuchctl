package templates

import (
	"html/template"
	"net/http"
)

func Render(w http.ResponseWriter, tmpl string) {
	render, err := template.ParseFiles(tmpl)
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
