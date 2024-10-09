package main

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"

	routes "filmPackager/internal/handlers"
	db "filmPackager/internal/store/db"
)

func main() {
  db.Connect()

	app := fiber.New()

	mux := routes.RegisterRoutes()
	log.Print("Listening on port 3000...")
	log.Fatal(http.ListenAndServe(":3000", mux))
}