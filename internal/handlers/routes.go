package routes

import (
	"context"
	access "filmPackager/internal/auth"
	"filmPackager/internal/store/db"
	"fmt"
	"html/template"
	"net/http"
	"net/mail"
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

// sets up the route multiplexer
func RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", IndexPage)
	mux.HandleFunc("/login/", GetLoginPage)
	mux.HandleFunc("/post-login/", PostLoginSubmit)
	mux.HandleFunc("/post-create/", PostCreateAccount)
	mux.HandleFunc("/get-create-account/", DirectToCreateAccount)
	mux.HandleFunc("/create-account/", GetCreateAccount)
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
		fmt.Println("no token cookie at the index")
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
		fmt.Println("token string passed to verify token: ",tokenString)
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
	fmt.Println("user Info: ", userInfo)
	data := HomeData{
		User: userInfo,
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html",
	"templates/doc-list.html", "templates/file-upload.html", "templates/sidebar.html", "templates/header.html",
	))
	// REPLACE THE NIL WITH DATA from DB
	fmt.Println("data: ", data)
	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")
	conn := db.Connect()
	defer conn.Close(context.Background())
	user, err := db.GetUser(conn, email, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid credentials")
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
	tmpl := template.Must(template.ParseFiles("templates/login-form.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
		w.Header().Set("HX-Redirect", "/")
		return
	}
}

func DirectToCreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
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
	IndexPage(w, r)
}