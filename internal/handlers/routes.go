package routes

import (
	"fmt"
	"html/template"
	"net/http"
)

// TEMP AUTH TEST FOR REDIRECT
var auth = false
func setAuthTrue () {
	auth = true
}

func Homepage(w http.ResponseWriter, r *http.Request) {
	// TEMP AUTH REDIRECT
	// redirect to false?
	if !auth {
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html",
	"templates/docList.html", "templates/fileupload.html", "templates/sidebar.html",
	))
	// REPLACE THE NIL WITH DATA from DB
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/loginFORM.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println("error executing the fucking template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	// Check if the request is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		setAuthTrue()
		// HTMX request, use HX-Redirect to tell HTMX to redirect
		w.Header().Set("HX-Redirect", "/")
		return
	}
}