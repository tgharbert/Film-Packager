package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"filmPackager/internal/application"
	docInf "filmPackager/internal/infrastructure/document"
	projectInf "filmPackager/internal/infrastructure/project"
	userInf "filmPackager/internal/infrastructure/user"

	"filmPackager/internal/interfaces"
	s3Conn "filmPackager/internal/store"
	db "filmPackager/internal/store/db"
)

func main() {
	db.PoolConnect()
	conn := db.GetPool()
	s3Client := s3Conn.GetS3Client(context.Background())
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatal("BUCKET env var not set")
	}
	engine := html.New("./views", ".html")
	if err := engine.Load(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/static", "./static")
	// wrap this up in a different file/struct??
	userRepo := userInf.NewPostgresUserRepository(conn)
	projectRepo := projectInf.NewPostgresProjectRepository(conn)
	docPGRepo := docInf.NewPostgresDocumentRepository(conn)
	docS3Repo := docInf.NewS3DocumentRepository(s3Client, bucket)
	userService := application.NewUserService(userRepo, projectRepo)
	projService := application.NewProjectService(projectRepo, docPGRepo)
	docService := application.NewDocumentService(docPGRepo, docS3Repo, userRepo)

	interfaces.RegisterRoutes(app, userService, projService, docService)

	log.Print("Listening on port 3000...")
	log.Fatal(app.Listen(":3000"))
}
