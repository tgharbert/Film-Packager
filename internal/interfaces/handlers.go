package interfaces

import (
	"filmPackager/internal/application"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, userService *application.UserService, projectService *application.ProjectService, documentService *application.DocumentService) {
	// app.Get("/", HomePage)
	// app.Get("/login/", GetLoginPage)
	app.Post("/post-login/", LoginUserHandler(userService))
	// app.Post("/post-create-account", PostCreateAccount)
	// app.Get("/get-create-account/", DirectToCreateAccount)
	// app.Get("/create-project/", CreateProject)
	// app.Get("/get-project/:id", GetProject)
	// app.Get("/logout/", Logout)
	app.Post("/file-submit/:project_id", UploadDocumentHandler(documentService))
	// app.Post("/search-users/:id", SearchUsers)
	// app.Post("/invite-member/:id/:project_id", InviteMember)
	// app.Post("/join-org/:project_id/:role", JoinOrg)
	// app.Get("/delete-project/:project_id/", DeleteOrg)
}

// user handlers:
func LoginUserHandler(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.FormValue("email"))
		password := strings.TrimSpace(c.FormValue("password"))
		if email == "" || password == "" {
			return c.Render("login-formHTML", fiber.Map{
				"Error": "Error: both fields must be filled!",
			})
		}
		user, err := svc.UserLogin(c.Context(), email, password)
		if err != nil {
			return c.Render("login-formHTML", fiber.Map{
				"Error": "Error: both fields must be filled!",
			})
		}
		tokenString, err := access.GenerateJWT(user.Id, user.Name, user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error generating JWT")
		}
		c.Cookie(&fiber.Cookie{
			Name:     "Authorization",
			Value:    "Bearer " + tokenString,
			HTTPOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(48 * time.Hour),
		})
		return c.Redirect("/")
	}
}

func UploadDocumentHandler(svc *application.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.FormValue("organization_id")
		fileType := c.FormValue("file_type")
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("File is required")
		}
		f, _ := file.Open()
		defer f.Close()
		orgIDInt, err := strconv.Atoi(orgID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		doc := &domain.Document{
			OrganizationID: orgIDInt,
			FileName:       file.Filename,
			FileType:       fileType,
		}
		err = svc.UploadDocument(c.Context(), doc, f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		// return the HTML fragment/template
		return c.JSON(doc)
	}
}
