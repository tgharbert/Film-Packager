package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	interfaces "filmPackager/internal/presentation"
)

func main() {
	log.Print("Starting server...")
	server := interfaces.NewServer(fiber.New())
	server.Start()
}
