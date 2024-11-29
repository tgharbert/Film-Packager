package interfaces

import (
	"filmPackager/internal/application"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain"
	"filmPackager/internal/store/db"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type HomeData struct {
	User *domain.User
	Orgs db.SelectProject
}

func RegisterRoutes(app *fiber.App, userService *application.UserService, projectService *application.ProjectService, documentService *application.DocumentService) {
	app.Get("/", GetHomePage(projectService))
	app.Get("/login/", GetLoginPage(userService))
	app.Post("/post-login/", LoginUserHandler(userService))
	// app.Post("/post-create-account", PostCreateAccount)
	app.Get("/get-create-account/", GetCreateAccount(userService))
	app.Get("/create-project/", CreateProject(projectService))
	// app.Get("/get-project/:id", GetProject)
	app.Get("/logout/", LogoutUser(userService))
	app.Post("/file-submit/:project_id", UploadDocumentHandler(documentService))
	// app.Post("/search-users/:id", SearchUsers)
	// app.Post("/invite-member/:id/:project_id", InviteMember)
	// app.Post("/join-org/:project_id/:role", JoinOrg)
	app.Get("/delete-project/:project_id/", DeleteProject(projectService))
}

// user handlers:
func GetLoginPage(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Render("login-form", nil)
		}
		tokenString = tokenString[len("Bearer "):]
		err := access.VerifyToken(tokenString)
		if err != nil {
			return c.Render("login-form", nil)
		}
		return c.Redirect("/")
	}
}

func GetCreateAccount(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("create-account", nil)
	}
}

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
			fmt.Println("error: ", err)
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

func LogoutUser(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "Authorization",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour), // Set expiration to the past to delete the cookie
			Path:     "/",                        // Ensure the path is the same as when the cookie was set
			HTTPOnly: true,                       // Ensure other flags match those of the original cookie
			Secure:   true,                       // Set to true if the original cookie was secure
		})
		return c.Redirect("/login/")
	}
}

func GetHomePage(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("Authorization")
		if c.Get("HX-Request") == "true" {
			c.Set("HX-Redirect", "/") // Redirect to homepage or desired URL
			return nil
		}
		if tokenString == "" {
			return c.Redirect("/login/")
		}
		tokenString = tokenString[len("Bearer "):]
		userInfo, err := access.GetUserNameFromToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}
		user, err := svc.GetUsersProjects(c.Context(), userInfo)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}
		return c.Render("index", *user)
	}
}

func CreateProject(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectName := c.FormValue("project-name")
		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}
		tokenString = tokenString[len("Bearer "):]
		userInfo, err := access.GetUserNameFromToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}
		project, err := svc.CreateNewProject(c.Context(), projectName, userInfo.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error creating org")
		}
		return c.Render("project-list-item", *project)
	}
}

func DeleteProject(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}
		projectId := c.Params("project_id")
		projIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		user, err := svc.DeleteProject(c.Context(), projIdInt, userInfo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting project")
		}
		// need to delete s3 items from bucket as well!
		fmt.Println("user info: ", user)
		return c.Render("project-list", fiber.Map{
			"Memberships": user.Memberships,
		})
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
