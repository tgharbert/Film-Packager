package interfaces

import (
	"context"
	"filmPackager/internal/application/authservice"
	"filmPackager/internal/application/commentservice"
	"filmPackager/internal/application/documentservice"
	"filmPackager/internal/application/membershipservice"
	"filmPackager/internal/application/middleware/auth"
	"filmPackager/internal/application/projectservice"
	"filmPackager/internal/application/userservice"
	commInf "filmPackager/internal/infrastructure/comment"
	docInf "filmPackager/internal/infrastructure/document"
	memInf "filmPackager/internal/infrastructure/membership"
	projectInf "filmPackager/internal/infrastructure/project"
	userInf "filmPackager/internal/infrastructure/user"
	"filmPackager/internal/presentation/routes"
	s3Conn "filmPackager/internal/store"
	"filmPackager/internal/store/db"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

type Server struct {
	fiberApp *fiber.App
}

func NewServer(app *fiber.App) *Server {
	// load the .env file locally if one exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// set up the S3 client
	s3Client := s3Conn.GetS3Client(context.Background())
	if s3Client == nil {
		log.Fatal("Error connecting to the s3 client")
	}
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatal("BUCKET env var not set")
	}

	// set up the database connection
	conn := db.PoolConnect()
	if conn == nil {
		log.Fatal("Error connecting to the database")
	}

	// set up views and static files
	viewEngine := html.New("./views", ".html")

	s := &Server{
		fiberApp: fiber.New(
			fiber.Config{
				Views: viewEngine,
				// 30MB file size limit
				BodyLimit: 1024 * 1024 * 30,
			},
		),
	}

	// serve the static files
	s.fiberApp.Static("/static", "./static")

	// instantiate the repositories
	userRepo := userInf.NewPostgresUserRepository(conn)
	projectRepo := projectInf.NewPostgresProjectRepository(conn)
	docPGRepo := docInf.NewPostgresDocumentRepository(conn)
	memberRepo := memInf.NewPostgresMembershipRepository(conn)
	docS3Repo := docInf.NewS3DocumentRepository(s3Client, bucket)
	commentRepo := commInf.NewPostgresCommentRepository(conn)

	// instantiate the services
	userService := userservice.NewUserService(userRepo, projectRepo)
	projService := projectservice.NewProjectService(projectRepo, docPGRepo, docS3Repo, userRepo, memberRepo, commentRepo)
	docService := documentservice.NewDocumentService(docPGRepo, docS3Repo, userRepo, memberRepo, projectRepo, commentRepo)
	memberService := membershipservice.NewMembershipService(memberRepo, userRepo)
	authService := authservice.NewAuthService(userRepo)
	commentService := commentservice.NewCommentService(commentRepo, userRepo)

	// register the middleware BEFORE registering the routes
	s.RegisterMiddleware(authService)

	// register the routes
	s.RegisterRoutes(userService, projService, docService, memberService, authService, commentService)

	return s
}

func (s *Server) RegisterMiddleware(authService *authservice.AuthService) {
	// add middleware here
	s.fiberApp.Use(
		requestid.New(
			requestid.Config{
				Generator: utils.UUIDv4,
			},
		),
	)

	s.fiberApp.Use(logger.New())
	s.fiberApp.Use(auth.New(authService))
}

func (s *Server) Start() error {
	return s.fiberApp.Listen("0.0.0.0:8080")
}

func (s *Server) RegisterRoutes(userService *userservice.UserService, projectService *projectservice.ProjectService, documentService *documentservice.DocumentService, membershipService *membershipservice.MembershipService, authService *authservice.AuthService, commentService *commentservice.CommentService) {
	// homepage
	s.fiberApp.Get("/", routes.GetHomePage(projectService))

	// auth routes - only for login
	s.fiberApp.Get("/login/", routes.GetLoginPage(authService))
	s.fiberApp.Post("/login/", routes.LoginUserHandler(authService))
	s.fiberApp.Post("/create-account", routes.PostCreateAccount(authService))

	// user routes
	s.fiberApp.Get("/create-account/", routes.GetCreateAccount(userService))
	s.fiberApp.Get("/logout/", routes.LogoutUser(userService))
	s.fiberApp.Get("/reset-password/", routes.GetResetPasswordPage(userService))
	s.fiberApp.Get("/verify-old-password/", routes.VerifyOldPassword(userService))
	s.fiberApp.Post("/new-password/", routes.SetNewPassword(userService))

	// member routes
	s.fiberApp.Post("/search-users/:id", routes.SearchMembersByName(membershipService))
	s.fiberApp.Post("/invite-member/:id/:project_id/", routes.InviteMember(membershipService))
	s.fiberApp.Get("/member/:project_id/:member_id/", routes.GetMemberPage(membershipService))
	s.fiberApp.Post("/member-roles/:project_id/:member_id/", routes.UpdateMemberRoles(membershipService))
	s.fiberApp.Get("/sidebar/:project_id/", routes.GetSidebar(membershipService))

	// project routes
	s.fiberApp.Get("/create-project/", routes.CreateProject(projectService))
	s.fiberApp.Post("/join-org/:project_id/:role", routes.JoinOrg(projectService))
	s.fiberApp.Get("/project/:project_id/", routes.GetProject(projectService))
	s.fiberApp.Get("/delete-project/:project_id/", routes.DeleteProject(projectService))
	s.fiberApp.Get("/click-delete-project/:project_id/", routes.ClickDeleteProject(projectService))
	s.fiberApp.Get("/cancel-delete-project/:project_id/", routes.CancelDeleteProject(projectService))
	s.fiberApp.Get("/project-name-form/:project_id/", routes.GetUpdateNameForm(projectService))
	s.fiberApp.Post("/project-name/:project_id/", routes.UpdateProjectName(projectService))

	// document routes
	s.fiberApp.Get("/doc-details/:doc_id", routes.GetDocDetails(documentService))
	s.fiberApp.Post("/file-submit/:project_id", routes.UploadDocumentHandler(documentService))
	s.fiberApp.Post("/lock-staged-docs/:project_id/", routes.LockStagedDocs(documentService))
	s.fiberApp.Get("/download-doc/:doc_id", routes.DownloadDocument(documentService))
	s.fiberApp.Get("/delete-doc/:doc_id", routes.DeleteDocument(documentService))
	s.fiberApp.Get("/preview-doc-page/:doc_id", routes.PreviewDocumentPage(documentService))
	s.fiberApp.Get("/preview-doc/:doc_id", routes.PreviewDocument(documentService))

	// comment routes
	s.fiberApp.Get("/doc-comments/:doc_id", routes.GetDocCommentSection(commentService))
	s.fiberApp.Post("/doc-comment/:doc_id", routes.AddDocComment(commentService))
	s.fiberApp.Delete("/doc-comment/:comment_id", routes.DeleteComment(commentService))
}
