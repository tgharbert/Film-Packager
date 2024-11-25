package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"filmPackager/internal/application"
	projectInf "filmPackager/internal/infrastructure/project"
	userInf "filmPackager/internal/infrastructure/user"

	// infrastructure "filmPackager/internal/infrastructure/project"
	"filmPackager/internal/interfaces"
	db "filmPackager/internal/store/db"
)

func main() {
	db.PoolConnect()
	conn := db.GetPool()
	// if err != nil {
	// 	panic("pool issues!")
	// }
	// defer conn.Release()
	engine := html.New("./templates", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/static", "./static")
	// routes.RegisterRoutes(app)
	userRepo := userInf.NewPostgresUserRepository(conn)
	projectRepo := projectInf.NewPostgresProjectRepository(conn)

	userService := application.NewUserService(userRepo, projectRepo)
	projService := application.NewProjectService(projectRepo)

	interfaces.RegisterRoutes(app, userService, projService, &application.DocumentService{})
	log.Print("Listening on port 3000...")
	log.Fatal(app.Listen(":3000"))
}
