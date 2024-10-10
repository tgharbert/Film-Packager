package routes

import (
	"context"
	access "filmPackager/internal/auth"
	"filmPackager/internal/store/db"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"golang.org/x/crypto/bcrypt"
)

type HomeData struct {
	User *access.UserInfo
	Orgs []db.Org
}

type Message struct {
	Error string
}

func RegisterRoutes(app *fiber.App) {
	app.Get("/", HomePage)
	app.Get("/login/", GetLoginPage)
	app.Post("/post-login/", PostLoginSubmit)
	app.Post("/post-create-account", PostCreateAccount)
	app.Get("/get-create-account/", DirectToCreateAccount)
	app.Get("/create-project/", CreateProject)
	app.Get("/get-project/:id", GetProject)
	app.Get("/logout/", Logout)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func HomePage(c *fiber.Ctx) error {
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
	conn := db.Connect()
	orgs, err := db.GetProjects(conn, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
	}
	data := HomeData{User: userInfo, Orgs: orgs,}
	return c.Render("index", data)
}

func PostLoginSubmit(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	if email == "" || password == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email or password cannot be empty")
	}
	conn := db.Connect()
	defer conn.Close(context.Background())
	user, err := db.GetUser(conn, email, password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving user")
	}
	if user.Password == "" {
		mess := Message{Error: "Incorrect Password"}
		return c.Render("login-error", mess) // Fiber automatically handles the template rendering
	}
	tokenString, err := access.GenerateJWT(user.Id, user.Name, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating JWT")
	}
	c.Cookie(&fiber.Cookie{
		Name:  "Authorization",
		Value: "Bearer " + tokenString,
		HTTPOnly: true,
		Path:  "/",
		Expires: time.Now().Add(48 * time.Hour),
	})
	return c.Redirect("/")
}

func GetLoginPage(c *fiber.Ctx) error {
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

func PostCreateAccount(c *fiber.Ctx) error {
	username := strings.Trim(c.FormValue("username"), " ")
	email := strings.Trim(c.FormValue("email"), " ")
	password := strings.Trim(c.FormValue("password"), " ")
	secondPassword := strings.Trim(c.FormValue("secondPassword"), " ")
	var mess Message
	if username == "" {
		mess.Error = "blank username"
		return c.Render("login-error", mess)
	}
	if email == "" {
		mess.Error = "blank email"
		return c.Render("login-error", mess)
	}
	if password != secondPassword {
		mess.Error = "passwords do not match"
		return c.Render("login-error", mess)
	}
	if len(password) < 6 || len(secondPassword) < 6 {
		mess.Error = "password is too short"
		return c.Render("login-error", mess)
	}
	if !isValidEmail(email) {
		mess.Error = "invalid email"
		return c.Render("login-error", mess)
	}
	conn := db.Connect()
	defer conn.Close(context.Background())
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error hashing password")
	}
	hashedStr := string(hash)
	user, err := db.CreateUser(conn, username, email, hashedStr)
	if err != nil {
		panic(err)
	}
	tokenString, err := access.GenerateJWT(user.Id, user.Name, user.Email, user.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error generating JWT")
	}
	c.Cookie(&fiber.Cookie{
		Name:  "Authorization",
		Value: "Bearer " + tokenString,
		HTTPOnly: true,
		Path:  "/",
		Expires: time.Now().Add(48 * time.Hour),
	})
	return c.Redirect("/")
}

func DirectToCreateAccount(c *fiber.Ctx) error {
	return c.Render("create-account", nil)
}

func Logout(c *fiber.Ctx) error {
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

func CreateProject(c *fiber.Ctx) error {
	// send the project list updated with new project data...
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
	conn := db.Connect()
	defer conn.Close(context.Background())
	org, err := db.CreateProject(conn, projectName, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error retrieving org")
	}
	return c.Render("project-list-item", org)
}

func GetProject(c *fiber.Ctx) error {
	// will need to get all the project information for a given project
	fmt.Println(c.Params("id"))
	// get the values for the project - personnel, docs, etc
	return nil
}