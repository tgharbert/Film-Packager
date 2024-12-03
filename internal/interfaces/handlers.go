package interfaces

import (
	"errors"
	"filmPackager/internal/application"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"filmPackager/internal/store/db"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type HomeData struct {
	User *user.User
	Orgs db.SelectProject
}

func RegisterRoutes(app *fiber.App, userService *application.UserService, projectService *application.ProjectService, documentService *application.DocumentService) {
	app.Get("/", GetHomePage(projectService))
	app.Get("/login/", GetLoginPage(userService))
	app.Post("/post-login/", LoginUserHandler(userService))
	app.Post("/post-create-account", PostCreateAccount(userService))
	app.Get("/get-create-account/", GetCreateAccount(userService))
	app.Get("/create-project/", CreateProject(projectService))
	app.Get("/get-project/:project_id", GetProject(projectService))
	app.Get("/logout/", LogoutUser(userService))
	app.Post("/file-submit/:project_id", UploadDocumentHandler(documentService))
	app.Post("/search-users/:id", SearchUsers(projectService))
	app.Post("/invite-member/:id/:project_id/", InviteMember(projectService))
	app.Post("/join-org/:project_id/:role", JoinOrg(projectService))
	app.Get("/delete-project/:project_id/", DeleteProject(projectService))
	app.Get("/get-member/:project_id/:member_id/", GetMemberPage(projectService))
	app.Post("/update-member-roles/:project_id/:member_id/", UpdateMemberRoles(projectService))
}

// user handlers:
func GetLoginPage(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// need to check if cookie is valid, if not render login
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

func PostCreateAccount(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		firstName := strings.Trim(c.FormValue("firstName"), " ")
		lastName := strings.Trim(c.FormValue("lastName"), " ")
		email := strings.Trim(c.FormValue("email"), " ")
		password := strings.Trim(c.FormValue("password"), " ")
		secondPassword := strings.Trim(c.FormValue("secondPassword"), " ")
		username := fmt.Sprintf("%s %s", firstName, lastName)
		var mess string
		if firstName == "" || lastName == "" {
			mess = "Error: please enter first and last name!"
			return c.Render("create-accountHTML", fiber.Map{
				"Error": mess,
			})
		}
		if email == "" {
			mess = "Error: email field left blank!"
			return c.Render("create-accountHTML", fiber.Map{
				"Error": mess,
			})
		}
		if password != secondPassword {
			mess = "Error: passwords do not match!"
			return c.Render("create-accountHTML", fiber.Map{
				"Error": mess,
			})
		}
		if len(password) < 6 || len(secondPassword) < 6 {
			mess = "Error: password need to be at least 6 characters!"
			return c.Render("create-accountHTML", fiber.Map{
				"Error": mess,
			})
		}
		if !user.IsValidEmail(email) {
			mess = "Error: invalid email address"
			return c.Render("create-accountHTML", fiber.Map{
				"Error": mess,
			})
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error hashing password")
		}
		hashedStr := string(hash)
		newUser := user.CreateNewUser(username, email, hashedStr)
		createdUser, err := svc.CreateUserAccount(c.Context(), newUser)
		if err != nil {
			if errors.Is(err, user.ErrUserAlreadyExists) {
				mess = "Error: user already exists!"
				return c.Render("create-accountHTML", fiber.Map{
					"Error": mess,
				})
			}
			return c.Status(fiber.StatusInternalServerError).SendString("error creating user")
		}
		tokenString, err := access.GenerateJWT(createdUser.Id, createdUser.Name, createdUser.Email)
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
		currentUser, err := svc.UserLogin(c.Context(), email, password)
		if err != nil {
			// send html with error message
			if errors.Is(err, user.ErrUserNotFound) {
				return c.Render("login-formHTML", fiber.Map{
					"Error": "Error: user not found!",
				})
			}
			return c.Status(fiber.StatusInternalServerError).SendString("error logging in")
		}
		tokenString, err := access.GenerateJWT(currentUser.Id, currentUser.Name, currentUser.Email)
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

func InviteMember(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Params("id")
		projectId := c.Params("project_id")
		projIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing project Id from request")
		}
		userIdInt, err := strconv.Atoi(userId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing user Id from request")
		}
		var errMess string
		invited, err := svc.InviteMember(c.Context(), projIdInt, userIdInt)
		if err != nil {
			if err == project.ErrMemberAlreadyInvited {
				// return the proper html fragment
				errMess = "You've already invited this user!"
				return c.Render("project-list", fiber.Map{
					"Error": errMess,
				})
			}
			return c.Status(fiber.StatusInternalServerError).SendString("error inviting user to project")
		}
		// return the proper html fragment - just the invited user...
		// should I clear the search results?
		return c.Render("invited-membersHTML", fiber.Map{
			"Invited": invited,
		})
	}
}

func LogoutUser(svc *application.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:  "Authorization",
			Value: "",
			// Set expiration to the past to delete the cookie
			Expires: time.Now().Add(-time.Hour),
			// Ensure the path is the same as when the cookie was set
			Path: "/",
			// Ensure other flags match those of the original cookie
			HTTPOnly: true,
			// Set to true if the original cookie was secure
			Secure: true,
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
		return c.Render("project-list", fiber.Map{
			"Memberships": user.Memberships,
		})
	}
}

func GetProject(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("project_id")
		projIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		project, err := svc.GetProjectDetails(c.Context(), projIdInt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error retrieving project data")
		}
		return c.Render("project-page", *project)
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
		doc := &document.Document{
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

func SearchUsers(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.FormValue("username")
		id := c.Params("id")
		users, err := svc.SearchForUsers(c.Context(), username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to query users")
		}
		return c.Render("search-resultsHTML", fiber.Map{
			"SearchedMembers": users,
			"ProjectId":       id,
		})
	}
}

func JoinOrg(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("project_id")
		// role := c.Params("role")
		projIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}
		user, err := svc.JoinProject(c.Context(), projIdInt, userInfo.Id)
		if err != nil {
			fmt.Println("error: ", err)
			return c.Status(fiber.StatusInternalServerError).SendString("error joining project")
		}
		return c.Render("selectOrgHTML", fiber.Map{"Memberships": user})
	}
}

func GetMemberPage(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		memberId := c.Params("member_id")
		memberIdInt, err := strconv.Atoi(memberId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		projectId := c.Params("project_id")
		projectIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		member, err := svc.GetProjectUser(c.Context(), projectIdInt, memberIdInt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting project user")
		}
		var availRoles []string
		allRoles := []string{"director", "producer", "writer", "cinematographer", "production_designer"}
		for _, role := range allRoles {
			if slices.Contains(member.Roles, role) {
				continue
			} else {
				availRoles = append(availRoles, role)
			}
		}
		return c.Render("member-detailsHTML", fiber.Map{"Member": *member, "ProjectId": projectId, "Roles": availRoles})
	}
}

func UpdateMemberRoles(svc *application.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		memberId := c.Params("member_id")
		memberIdInt, err := strconv.Atoi(memberId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}
		projectId := c.Params("project_id")
		projectIdInt, err := strconv.Atoi(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		role := c.FormValue("role-select")
		member, err := svc.UpdateMemberRoles(c.Context(), projectIdInt, memberIdInt, userInfo.Id, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error updating member roles")
		}
		var availRoles []string
		allRoles := []string{"director", "producer", "writer", "cinematographer", "production_designer"}
		for _, role := range allRoles {
			if slices.Contains(member.Roles, role) {
				continue
			} else {
				availRoles = append(availRoles, role)
			}
		}
		return c.Render("member-detailsHTML", fiber.Map{"Member": *member, "ProjectId": projectId, "Roles": availRoles})
	}
}
