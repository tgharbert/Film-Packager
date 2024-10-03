package main

import (
	"log"
	"net/http"

	routes "filmPackager/internal/handlers"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", routes.Homepage)
	http.HandleFunc("/login/", routes.Login)
	http.HandleFunc("/post-login/", routes.PostLogin)
	log.Print("Listening on port 3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}