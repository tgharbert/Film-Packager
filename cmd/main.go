package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	interfaces "filmPackager/internal/presentation"
)

func main() {
	server := interfaces.NewServer(fiber.New())
	log.Print("Listening on port 8080... I hope")
	log.Fatal(server.Start())
}
