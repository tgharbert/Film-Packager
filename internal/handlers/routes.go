package routes

import (
	"context"
	"filmPackager/internal/store/db"
	"fmt"
	"html/template"
	"net/http"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

// TEMP AUTH TEST FOR REDIRECT
var auth = false
func setAuthTrue () {
	auth = true
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// sets up the route multiplexer
func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", Homepage)
	mux.HandleFunc("/login/", GetLoginPage)
	mux.HandleFunc("/post-login/", PostLoginSubmit)
	mux.HandleFunc("/post-create/", PostCreateAccount)
	mux.HandleFunc("/get-create-account/", DirectToCreateAccount)
	mux.HandleFunc("/create-account/", GetCreateAccount)
	return mux
}

func Homepage(w http.ResponseWriter, r *http.Request) {
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

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login-form.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	fmt.Println(email, password)
	conn := db.Connect()
	defer conn.Close(context.Background())
	user, err := db.GetUser(conn, email, password)
	if err != nil {
		fmt.Println("HIT THIS PANIC")
		// return that no user is found, please check email and pw
		panic(err)
	}
	fmt.Println("logged in: ", user)
	if r.Header.Get("HX-Request") == "true" {
		setAuthTrue()
		// HTMX request, use HX-Redirect to tell HTMX to redirect
		w.Header().Set("HX-Redirect", "/")
		return
	}
	// get the user with the email and hash...
}

func PostCreateAccount(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	secondPassword := r.PostFormValue("secondPassword")
	if password != secondPassword {
		fmt.Println("password and second password do not match!")
		// return appropriate html...
	}
	if !isValidEmail(email) {
		fmt.Println("email is not valid!")
		// return appropriate email html...
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
	fmt.Println("created user: ", user)
	if err != nil {
		panic(err)
	}
	if r.Header.Get("HX-Request") == "true" {
		setAuthTrue()
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
