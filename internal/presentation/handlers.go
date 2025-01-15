package interfaces

import (
	"errors"
	"filmPackager/internal/application/membershipservice"
	"filmPackager/internal/application/projectservice"
	"filmPackager/internal/application/userservice"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// user handlers:
func GetLoginPage(svc *userservice.UserService) fiber.Handler {
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

func PostCreateAccount(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		firstName := strings.Trim(c.FormValue("firstName"), " ")
		lastName := strings.Trim(c.FormValue("lastName"), " ")
		email := strings.Trim(c.FormValue("email"), " ")
		password := strings.Trim(c.FormValue("password"), " ")
		secondPassword := strings.Trim(c.FormValue("secondPassword"), " ")
		username := fmt.Sprintf("%s %s", firstName, lastName)
		var mess string
		// TODO: I want to move all of this into the application layer and wrap in a util function
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
		createdUser, err := svc.CreateUser(c.Context(), username, email, password)
		// newUser := user.CreateNewUser(username, email, hashedStr)
		//createdUser, err := svc.CreateUserAccount(c.Context(), newUser)
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

func GetCreateAccount(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("create-account", nil)
	}
}

func LoginUserHandler(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: push these checks to the application layer
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

func InviteMember(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Params("id")
		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing project Id from request")
		}

		userUUID, err := uuid.Parse(userId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing user Id from request")
		}

		var errMess string

		invitedMembers, err := svc.InviteUserToProject(c.Context(), userUUID, projUUID)
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

		return c.Render("form-search-membersHTML", fiber.Map{
			"Invited": invitedMembers,
		})
	}
}

func LogoutUser(svc *userservice.UserService) fiber.Handler {
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

func GetHomePage(svc *projectservice.ProjectService) fiber.Handler {
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
		rv, err := svc.GetUsersProjects(c.Context(), userInfo)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}
		return c.Render("index", *rv)
	}
}

func SearchUsers(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		searchTerm := c.FormValue("username")
		projectID := c.Params("id")

		// parse the project id
		projUUID, err := uuid.Parse(projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// search for new members
		users, err := svc.SearchForNewMembers(c.Context(), searchTerm, projUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to query users")
		}

		return c.Render("search-resultsHTML", fiber.Map{
			"SearchedMembers": users,
			"ProjectID":       projectID,
		})
	}
}

func JoinOrg(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}

		err = svc.JoinProject(c.Context(), projUUID, userInfo.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error joining project")
		}

		rv, err := svc.GetUsersProjects(c.Context(), userInfo)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}

		return c.Render("selectOrgHTML", *rv)
	}
}

func GetMemberPage(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		memberId := c.Params("member_id")
		memberUUID, err := uuid.Parse(memberId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.GetMembership(c.Context(), projUUID, memberUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting project member")
		}

		return c.Render("member-detailsHTML", fiber.Map{"Member": *rv.Membership, "ProjectId": projectId, "Roles": rv.AvailableRoles})
	}
}

func UpdateMemberRoles(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		mUserID := c.Params("member_id")
		mUserUUID, err := uuid.Parse(mUserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		role := c.FormValue("role-select")

		member, err := svc.UpdateMemberRoles(c.Context(), projUUID, mUserUUID, role)
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

func GetSidebar(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pIDString := c.Params("project_id")
		pID, err := uuid.Parse(pIDString)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// get all project membership info for the sidebar
		rv, err := svc.GetProjectMemberships(c.Context(), pID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting project memberships")
		}

		// confirm render and target, something seems off right now, likely with HTMX
		return c.Render("sidebarHTML", fiber.Map{
			"ProjectID": pID,
			"Invited":   rv.Invited,
			"Members":   rv.Members,
		})
	}
}
