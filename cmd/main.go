package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"filmPackager/internal/application"
	projectInf "filmPackager/internal/infrastructure/project"
	userInf "filmPackager/internal/infrastructure/user"

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
	engine := html.New("./views", ".html")
	if err := engine.Load(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/static", "./static")

	// routes.RegisterRoutes(app)
	// wrap this up in a different file/struct??
	userRepo := userInf.NewPostgresUserRepository(conn)
	projectRepo := projectInf.NewPostgresProjectRepository(conn)
	// docPGRepo := docInf.NewPostgresDocumentRepository(conn)
	// docS3Repo := docInf.NewS3DocumentRepository(&s3.Client{}, bucket)
	userService := application.NewUserService(userRepo, projectRepo)
	projService := application.NewProjectService(projectRepo)
	// docService := application.NewDocumentService(docPGRepo, docS3Repo)

	interfaces.RegisterRoutes(app, userService, projService, &application.DocumentService{})
	log.Print("Listening on port 3000...")
	log.Fatal(app.Listen(":3000"))
}
