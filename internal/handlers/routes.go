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
	if !auth {
		http.Redirect(w, r, "/login/", http.StatusFound)
	}
	tmpl := template.Must(template.ParseFiles("templates/index.html",
	"templates/doc-list.html", "templates/file-upload.html", "templates/sidebar.html",
	))
	// REPLACE THE NIL WITH DATA from DB
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login-form.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println("error executing the fucking template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DirectHome(w http.ResponseWriter, r *http.Request) {
	// Here we will check if the user has an account, if they don't then sign up?
	// Check if the request is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		setAuthTrue()
		// HTMX request, use HX-Redirect to tell HTMX to redirect
		w.Header().Set("HX-Redirect", "/")
		return
	}
}

func DirectToCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		setAuthTrue()
		// HTMX request, use HX-Redirect to tell HTMX to redirect
		w.Header().Set("HX-Redirect", "/create-account/")
		return
	}
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/create-account.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}