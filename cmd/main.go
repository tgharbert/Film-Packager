package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	interfaces "filmPackager/internal/presentation"
)

func main() {
	fmt.Println("Hello, World!")
	server := interfaces.NewServer(fiber.New())
	log.Print("Listening on port 8080... I hope")
	server.Start()
}
