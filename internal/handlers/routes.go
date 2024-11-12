package routes

import (
	access "filmPackager/internal/auth"
	"filmPackager/internal/store/db"
	"fmt"
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
		fmt.Println(err)
		return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
	}
	// fmt.Println(orgs.Pending.Roles[0])
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
		fmt.Println("error here", err)
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
		s3Client := getS3Session()
		err = godotenv.Load()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error loading .env file")
		}
		bucket := os.Getenv("S3_BUCKET_NAME")
		err = DeleteMultipleS3Objects(s3Client, bucket, keys)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting file from bucket")
		}
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

func getS3Session() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	return s3.New(sess)
}

func DeleteMultipleS3Objects(s3Client *s3.S3, bucket string, keys []string) error {
	objectsToDelete := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objectsToDelete[i] = &s3.ObjectIdentifier{
			Key: aws.String(key),
		}
	}
	input := &s3.DeleteObjectsInput{
		Bucket : aws.String(bucket),
		Delete: &s3.Delete{
			Objects: objectsToDelete,
			Quiet: aws.Bool(true),
		},
	}
	output, err := s3Client.DeleteObjects(input)
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}
	for _, deleted := range output.Deleted {
		log.Printf("Deleted object: %s", *deleted.Key)
	}
	return nil
}

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

	s3Client := getS3Session()
	err = godotenv.Load()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("error loading .env file")
	}
	bucket := os.Getenv("S3_BUCKET_NAME")
	key := file.Filename

	// EXTRACT THE BELOW INTO A SEPARATE FUNC
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
		Body: f,
	})
	if err != nil {
		log.Printf("Error uploading file: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to upload file to s3")
	}
	// publicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	// Should this return anything for the HTML?
	// also should I modify this query to limit the amount of documents that I'm storing
	err = db.SaveDocument(db.DBPool, projIdInt, key, userInfo.Id, fileType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed save file info to db")
	}
	// this will need to return the correct HTML with the data that I'm looking for
	return nil
}