package routes

import (
	"html/template"
	"net/http"
)

func Homepage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html", "templates/sidebar.html", "templates/docList.html", "templates/fileUpload.html"))

	// REPLACE THE NIL WITH DATA
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
