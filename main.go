package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	routes "filmPackager/internal/handlers"
	db "filmPackager/internal/store/db"
)

func main() {
  // db.Connect()
	db.PoolConnect()
	conn, err := db.GetPool().Acquire(context.Background())
	if err != nil {
		panic("pool issues!")
	}
	defer conn.Release()
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/static", "./static")
	routes.RegisterRoutes(app)
	log.Print("Listening on port 3000...")
	log.Fatal(app.Listen(":3000"))
}