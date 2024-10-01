package main

import (
	"log"
	"net/http"

	routes "filmPackager/internal/handlers"
)


func main() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", routes.Homepage)
	log.Print("Listening on port 9090...")
	log.Fatal(http.ListenAndServe(":9090", nil))
}