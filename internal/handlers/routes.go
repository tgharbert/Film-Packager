package routes

import (
	"context"
	"filmPackager/internal/store/db"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// TEMP AUTH TEST FOR REDIRECT
var auth = false
func setAuthTrue () {
	auth = true
}

// sets up the route multiplexer
func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", Homepage)
	mux.HandleFunc("/login/", Login)
	mux.HandleFunc("/post-login/", DirectHome)
	mux.HandleFunc("/get-create-account/", DirectToCreateAccount)
	mux.HandleFunc("/create-account/", GetCreateAccount)
	return mux
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

	// db.Connect()
	// get the password and the email from the template,
	// pass them to the function to get the user, if the user password doesn't equal the hashed pw, kill, else proceed
	if err != nil {
		fmt.Println("error executing the fucking template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func DirectHome(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	secondPassword := r.PostFormValue("secondPassword")
	if password != secondPassword {
		fmt.Println("email and second email do not match!")
		// return appropriate html...
	}
	conn := db.Connect()
	defer conn.Close(context.Background())
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("error with password")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	hashedStr := string(hash)
	user, err := db.CreateUser(conn, username, email, hashedStr)
	// SEND THE USER WITH THE HTML
	fmt.Println(user)
	if err != nil {
		panic(err)
	}
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

func GetCreateAccount(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/create-account.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}