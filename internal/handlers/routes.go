package routes

import (
	access "filmPackager/internal/auth"
	s3Conn "filmPackager/internal/store"
	"filmPackager/internal/store/db"
	"fmt"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

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
	app.Post("/file-submit/:project_id", PostDocument)
	app.Post("/search-users/:id", SearchUsers)
	app.Post("/invite-member/:id/:project_id", InviteMember)
	app.Post("/join-org/:project_id/:role", JoinOrg)
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

// AUTH ROUTES
func PostLoginSubmit(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	var mess Message
	if email == "" || password == "" {
		mess.Error = "Error: both fields must be filled!"
		return c.Render("login-formHTML", mess)
	}
	user, err := db.GetUser(db.DBPool, email, password)
	if err != nil {
		mess.Error = "Error: cannot find user, please verify correct login!"
		return c.Render("login-formHTML", mess)
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
	err = db.CreateProject(db.DBPool, projectName, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error creating org")
	}
	orgs, err := db.GetProjects(db.DBPool, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error retrieving all orgs")
	}
	return c.Render("selectOrgHTML" , fiber.Map{
		"Orgs": orgs,
	})
}

// SHOULD RENDER THE HTML BASED ON ROLE??
// USERS SHOULD SEE THE INVITED BUT NOT BE ABLE TO SEARCH MEMBERS UNLESS PROD, DIR, OR OWNER
// how to do this??
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

// JOINING AND INVITING TO ORGS WORK
func SearchUsers(c *fiber.Ctx) error {
	username := c.FormValue("username")
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
	tokenString := c.Cookies("Authorization")
	if tokenString == "" {
		return c.Redirect("/login/")
	}
	tokenString = tokenString[len("Bearer "):]
	userInfo, err := access.GetUserNameFromToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
	}
	projectId := c.Params("project_id")
	role := c.Params("role")
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	err = db.JoinOrg(db.DBPool, projIdInt, userInfo.Id, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error querying database")
	}
	orgs, err := db.GetProjects(db.DBPool, userInfo.Id)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
	}
	return c.Render("selectOrgHTML" , fiber.Map{
		"Orgs": orgs,
	})
}

func DeleteOrg(c *fiber.Ctx) error {
	userInfo, err := access.GetUserDataFromCookie(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
	}
	projectId := c.Params("project_id")
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	keys, err := db.GetDocKeysForOrgDelete(db.DBPool, projIdInt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error retrieving keys")
	}
	if len(keys) != 0 {
		s3Client := s3Conn.GetS3Session()
		err = godotenv.Load()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error loading .env file")
		}
		bucket := os.Getenv("S3_BUCKET_NAME")
		err = s3Conn.DeleteMultipleS3Objects(s3Client, bucket, keys)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting file from bucket")
		}
	}
	err = db.DeleteOrg(db.DBPool, projIdInt)
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

// what should happen here is that we query the database to check for an already staged file
// if this filetype then if one exists delete it from the s3 and then delete it from the database
	// then write the new data to the database - continue with update db method??
// otherwise simply write the file to the s3 and then write to the db
func PostDocument(c *fiber.Ctx) error {
	projectId := c.Params("project_id")
	projIdInt, err := strconv.Atoi(projectId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
	}
	tokenString := c.Cookies("Authorization")
	if tokenString == "" {
		return c.Redirect("/login/")
	}
	tokenString = tokenString[len("Bearer "):]
	userInfo, err := access.GetUserNameFromToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
	}
	file, err := c.FormFile("file")

	// 25MB file limit -- matches what GMail allows as an attachment
	// at the moment I'm not hitting this with larger files...
	if file.Size > 25 * 1024 * 1024 {
		fmt.Println("hit the test limiter")
		// MODIFY to send HTML ERROR
		return fmt.Errorf("file too large: %v", err)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error uploading file")
	}
	fileType := c.FormValue("file-type")
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to open file")
	}
	defer f.Close()

	s3Client := s3Conn.GetS3Session()
	err = godotenv.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error loading .env file")
	}
	bucket := os.Getenv("S3_BUCKET_NAME")
	key := file.Filename

	oldFile, err := db.CheckForStagedDoc(db.DBPool, projIdInt, fileType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error checking for staged doc")
	} else if oldFile == "" {
		err = s3Conn.WriteToS3(s3Client, bucket, key, f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to upload file to s3")
		}
		err = db.SaveDocument(db.DBPool, projIdInt, key, userInfo.Id, fileType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed save file info to db")
		}
	} else {
		err := s3Conn.DeleteS3Object(s3Client, bucket, oldFile)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete s3 file")
		}
		err = s3Conn.WriteToS3(s3Client, bucket, key, f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to upload file to s3")
		}
		err = db.OverWriteDoc(db.DBPool, projIdInt, key, userInfo.Id, fileType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to update doc in db")
		}
	}
	// also should I modify this query to limit the amount of documents that I'm storing
	// this will need to return the correct HTML with the data that I'm looking for
	return nil
}