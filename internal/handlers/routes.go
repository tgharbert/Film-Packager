package routes

import (
	access "filmPackager/internal/auth"
	"filmPackager/internal/store/db"
	"fmt"
	"log"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber/v2"

	"golang.org/x/crypto/bcrypt"
)

type HomeData struct {
	User *access.UserInfo
	Orgs db.SelectProject
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
	app.Post("/file-submit/", PostDocument)
	app.Post("/search-users/:id", SearchUsers)
	app.Post("/invite-member/:id/:project_id", InviteMember)
	app.Post("/join-org/:id/:project_id/:role", JoinOrg)
	app.Get("/delete-project/:project_id/", DeleteOrg)
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
	orgs, err := db.GetProjects(db.DBPool, userInfo.Id)
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
	user, err := db.GetUser(db.DBPool, email, password)
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
	firstName := strings.Trim(c.FormValue("firstName"), " ")
	lastName := strings.Trim(c.FormValue("lastName"), " ")
	email := strings.Trim(c.FormValue("email"), " ")
	password := strings.Trim(c.FormValue("password"), " ")
	secondPassword := strings.Trim(c.FormValue("secondPassword"), " ")
	// concat the first and last names
	username := fmt.Sprintf("%s %s", firstName, lastName)
	var mess Message
	if firstName == "" || lastName == "" {
		mess.Error = "Error: please enter first and last name!"
		return c.Render("create-accountHTML", mess)
	}
	if email == "" {
		mess.Error = "Error: email field left blank!"
		return c.Render("create-accountHTML", mess)
	}
	if password != secondPassword {
		mess.Error = "Error: passwords do not match!"
		return c.Render("create-accountHTML", mess)
	}
	if len(password) < 6 || len(secondPassword) < 6 {
		mess.Error = "Error: password need to be at least 6 characters!"
		return c.Render("create-accountHTML", mess)
	}
	if !isValidEmail(email) {
		mess.Error = "Error: invalid email address"
		return c.Render("create-accountHTML", mess)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error hashing password")
	}
	hashedStr := string(hash)
	user, err := db.CreateUser(db.DBPool, username, email, hashedStr)
	if err != nil {
		mess.Error = "Error: user already exists with this email!"
		return c.Render("create-accountHTML", mess)
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
	org, err := db.CreateProject(db.DBPool, projectName, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error retrieving org")
	}
	return c.Render("project-list-item", org)
}

func GetProject(c *fiber.Ctx) error {
	id := c.Params("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	projectPageData, err := db.GetProjectPageData(db.DBPool, idInt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error retriving project information")
	}
	return c.Render("project-page", projectPageData)
}


func SearchUsers(c *fiber.Ctx) error {
	username := c.FormValue("username")
	// MODIFY - GET USER ID FROM COOKIE?
	id := c.Params("id")
	users, err := db.SearchForUsers(db.DBPool, username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to query users")
	}
	return c.Render("search-resultsHTML", fiber.Map{
		"ProjectId": id,
		"FoundUsers": users,
	})
}

func InviteMember(c *fiber.Ctx) error {
	// MODIFY - GET USER ID FROM COOKIE?
	memberId := c.Params("id")
	role := c.FormValue("role")
	projectId := c.Params("project_id")
	memIdInt, err := strconv.Atoi(memberId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	users, err := db.InviteUserToOrg(db.DBPool, memIdInt, projIdInt, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error adding user to db")
	}
	return c.Render("new-list-of-invitesHTML", fiber.Map{
		"Members": users,
	})
}

func JoinOrg(c *fiber.Ctx) error {
	// MODIFY - GET USER ID FROM COOKIE?
	userId := c.Params("id")
	projectId := c.Params("project_id")
	role := c.Params("role")
	id, err := strconv.Atoi(userId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	err = db.JoinOrg(db.DBPool, projIdInt, id, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error querying database")
	}
	orgs, err := db.GetProjects(db.DBPool, id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
	}
	return c.Render("selectOrgHTML" , fiber.Map{
		"Orgs": orgs,
	})
}

func DeleteOrg(c *fiber.Ctx) error {
	// getting the user info from the cookie - should be done everywhere?
	userInfo, err := access.GetUserDataFromCookie(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
	}
	projectId := c.Params("project_id")
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	err = db.DeleteOrg(db.DBPool, projIdInt, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error querying the database")
	}
	orgs, err := db.GetProjects(db.DBPool, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
	}
	return c.Render("selectOrgHTML" , fiber.Map{
		"Orgs": orgs,
	})
}

// PROJECT DOCUMENT WORK
func getS3Session() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"), // replace with my region
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	return s3.New(sess)
}

func PostDocument(c *fiber.Ctx) error {
	// hit the thang?
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error uploading file")
	}
	// FIX THIS LATER TO STORE THE VALUES
	// fileType := c.FormValue("file-type")
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to open file")
	}
	defer f.Close()

	// intialize the s3 client
	s3Client := getS3Session()
	bucket := "bucket-name" // replace with my bucket name
	key := file.Filename

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
		Body: f,
		ACL: aws.String("public-read"),
	})
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to upload file to s3")
	}
	return nil
}