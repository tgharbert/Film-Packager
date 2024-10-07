package main

import (
	"log"
	"net/http"

	routes "filmPackager/internal/handlers"
	db "filmPackager/internal/store/db"
)

func main() {
  db.Connect()
	// acct := access.CreateAccount("tgharbert3@gmail.com", )
	mux := routes.RegisterRoutes()
	log.Print("Listening on port 3000...")
	log.Fatal(http.ListenAndServe(":3000", mux))
}