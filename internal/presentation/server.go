package interfaces

import (
	"context"
	"filmPackager/internal/application/documentservice"
	"filmPackager/internal/application/membershipservice"
	"filmPackager/internal/application/projectservice"
	"filmPackager/internal/application/userservice"
	docInf "filmPackager/internal/infrastructure/document"
	memInf "filmPackager/internal/infrastructure/membership"
	projectInf "filmPackager/internal/infrastructure/project"
	userInf "filmPackager/internal/infrastructure/user"
	s3Conn "filmPackager/internal/store"
	"filmPackager/internal/store/db"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Server struct {
	fiberApp *fiber.App
	// embed the auth service here eventually
}

func NewServer(app *fiber.App) *Server {
	// set up the database connection
	db.PoolConnect()
	conn := db.GetPool()

	// set up the S3 client
	s3Client := s3Conn.GetS3Client(context.Background())
	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		log.Fatal("BUCKET env var not set")
	}

	// set up views and static files
	viewEngine := html.New("./views", ".html")

	s := &Server{
		fiberApp: fiber.New(
			fiber.Config{
				Views: viewEngine,
			},
		),
	}

	s.fiberApp.Static("/static", "./static")
	// instantiate the repositories
	userRepo := userInf.NewPostgresUserRepository(conn)
	projectRepo := projectInf.NewPostgresProjectRepository(conn)
	docPGRepo := docInf.NewPostgresDocumentRepository(conn)
	memberRepo := memInf.NewPostgresMembershipRepository(conn)
	docS3Repo := docInf.NewS3DocumentRepository(s3Client, bucket)

	// instantiate the services
	userService := userservice.NewUserService(userRepo, projectRepo)
	projService := projectservice.NewProjectService(projectRepo, docPGRepo, userRepo, memberRepo)
	docService := documentservice.NewDocumentService(docPGRepo, docS3Repo, userRepo)
	memberService := membershipservice.NewMembershipService(memberRepo, userRepo)

	// register the routes
	s.RegisterRoutes(userService, projService, docService, memberService)

	return s
}

func (s *Server) Start() error {
	return s.fiberApp.Listen(":3000")
}

func (s *Server) RegisterRoutes(userService *userservice.UserService, projectService *projectservice.ProjectService, documentService *documentservice.DocumentService, membershipService *membershipservice.MembershipService) {
	s.fiberApp.Get("/", GetHomePage(projectService))
	s.fiberApp.Get("/login/", GetLoginPage(userService))
	s.fiberApp.Post("/post-login/", LoginUserHandler(userService))
	s.fiberApp.Post("/post-create-account", PostCreateAccount(userService))
	s.fiberApp.Get("/get-create-account/", GetCreateAccount(userService))
	s.fiberApp.Get("/create-project/", CreateProject(projectService))
	s.fiberApp.Get("/get-project/:project_id/", GetProject(projectService))
	s.fiberApp.Get("/logout/", LogoutUser(userService))
	s.fiberApp.Post("/file-submit/:project_id", UploadDocumentHandler(documentService))
	s.fiberApp.Post("/search-users/:id", SearchUsers(membershipService))
	s.fiberApp.Post("/invite-member/:id/:project_id/", InviteMember(membershipService))
	s.fiberApp.Post("/join-org/:project_id/:role", JoinOrg(projectService))
	s.fiberApp.Get("/delete-project/:project_id/", DeleteProject(projectService))
	s.fiberApp.Get("/get-member/:project_id/:member_id/", GetMemberPage(membershipService))
	s.fiberApp.Post("/update-member-roles/:project_id/:member_id/", UpdateMemberRoles(membershipService))
	s.fiberApp.Get("/get-doc-details/:doc_id", GetDocDetails(documentService))
	s.fiberApp.Get("/get-sidebar/:project_id/", GetSidebar(membershipService))
}
