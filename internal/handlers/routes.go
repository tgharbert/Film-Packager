package routes

import (
	"context"
	access "filmPackager/internal/auth"
	"filmPackager/internal/store/db"
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name string
	Email string
	Role string
}

type HomeData struct {
	User *access.UserInfo
}

type Message struct {
	Error string
}

// sets up the route multiplexer
func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", IndexPage)
	mux.HandleFunc("/login/", GetLoginPage)
	mux.HandleFunc("/post-login/", PostLoginSubmit)
	mux.HandleFunc("/post-create/", PostCreateAccount)
	mux.HandleFunc("/get-create-account/", DirectToCreateAccount)
	mux.HandleFunc("/logout/", Logout)
	return mux
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	// Retrieve JWT from the "Authorization" cookie
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		GetLoginPage(w, r) // Redirect to login page if cookie is missing
		return
	}

		// Extract the JWT token from the cookie value
	tokenString := cookie.Value[len("Bearer "):]

	if tokenString == "" {
		GetLoginPage(w, r)
		return
	}
	err = access.VerifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		GetLoginPage(w, r)
		return
	}
	HomePage(w, r)
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	// Retrieve JWT from the "Authorization" cookie
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		// Redirect to login page if cookie is missing
		GetLoginPage(w, r)
		return
	}
	// Extract the JWT token from the cookie value
	tokenString := cookie.Value[len("Bearer "):]
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}

	err = access.VerifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}
	userInfo, err := access.GetUserNameFromToken(tokenString)
	if err != nil {
		fmt.Errorf("Failed extracting user from token: %v", err)
		return
	}
	data := HomeData{
		User: userInfo,
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html",
	"templates/doc-list.html", "templates/file-upload.html", "templates/sidebar.html", "templates/header.html",
	))
	// REPLACE THE NIL WITH DATA from DB
	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := strings.Trim(r.PostFormValue("email"), " ")
	password := strings.Trim(r.PostFormValue("password"), " ")
	if email == "" || password == "" {
		return
	}
	conn := db.Connect()
	defer conn.Close(context.Background())
	user, err := db.GetUser(conn, email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Error retrieving user")
	}
	var mess Message
	if user.Password == "" {
		fmt.Println("HERE")
		mess.Error = "Incorrect Password"
		tmpl := template.Must(template.ParseFiles("templates/login-error.html"))
		err := tmpl.ExecuteTemplate(w, "loginErrorHTML", mess)
		if err != nil {
			return
			// http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
		// w.Header.Set("")
	}
	// if err != nil {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	fmt.Fprint(w, "Invalid credentials")

	// }
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		// return that no user is found, please check email and pw
		// panic(err)
		return
	}
	tokenString, err := access.GenerateJWT(user.Name, user.Email, user.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error generating JWT")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: "Authorization",
		Value: "Bearer " + tokenString,
		HttpOnly: true,
		Path: "/",
	})
	if r.Header.Get("HX-Request") == "true" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

func GetLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/login-form.html", "templates/login-form-template.html", "templates/create-account-template.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostCreateAccount(w http.ResponseWriter, r *http.Request) {
	username := strings.Trim(r.PostFormValue("username"), " ")
	email := strings.Trim(r.PostFormValue("email"), " ")
	password := strings.Trim(r.PostFormValue("password"), " ")
	secondPassword := strings.Trim(r.PostFormValue("secondPassword"), " ")
	if username == "" {
		return
	}
	if email == "" {
		return
	}
	if password != secondPassword {
		fmt.Println("password and second password do not match!")
		// return appropriate html...
	}
	if len(password) < 6 || len(secondPassword) < 6 {
		fmt.Println("password is too short!")
	}
	if !isValidEmail(email) {
		fmt.Println("email is not valid!")
		// return appropriate email html...
		return
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
		w.Header().Set("HX-Redirect", "/")
		return
	}
}

func DirectToCreateAccount(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/create-account-template.html"))
	err := tmpl.ExecuteTemplate(w, "create-accountHTML", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name: "Authorization",
		Value: "",
		Expires: time.Unix(0, 0),
		Path: "/",
		MaxAge: -1,
		HttpOnly: true,
		Secure: true,
	})
	GetLoginPage(w, r)
}