package routes

import (
	"errors"
	"filmPackager/internal/application/userservice"
	access "filmPackager/internal/auth"
	"filmPackager/internal/domain/user"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

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
			} else if errors.Is(err, user.ErrInvalidPassword) {
				return c.Render("login-formHTML", fiber.Map{
					"Error": "Error: invalid password!",
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

func GetResetPasswordPage(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the user from the cookie
		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		return c.Render("reset-passwordHTML", nil)
	}
}

func VerifyOldPassword(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}
		tokenString = tokenString[len("Bearer "):]

		userInfo, err := access.GetUserNameFromToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}

		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("password1"))
		pw2 := strings.TrimSpace(c.FormValue("password2"))

		if pw1 == "" || pw2 == "" {
			return c.Render("login-formHTML", fiber.Map{
				"Error": "Error: both fields must be filled!",
			})
		}

		// verify that pw1 and pw2 are the same
		if pw1 != pw2 {
			return c.Render("reset-passwordHTML", fiber.Map{
				"Error": "Error: passwords do not match!",
			})
		}

		// verify that the pw is correct
		err = svc.VerifyOldPassword(c.Context(), userInfo.Id, pw1)
		if err != nil {
			return c.Render("reset-passwordHTML", fiber.Map{
				"Error": "Error: Incorrect password!",
			})
		}

		return c.Render("new-pw-formHTML", userInfo)
	}
}

func SetNewPassword(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		u, err := access.GetUserNameFromToken(tokenString)

		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("new-password1"))
		pw2 := strings.TrimSpace(c.FormValue("new-password2"))

		if pw1 != pw2 {
			return c.Render("new-pw-formHTML", fiber.Map{
				"Error": "Error: passwords do not match!",
			})
		}

		err = svc.SetNewPassword(c.Context(), u.Id, pw1)
		if err != nil {
			return c.Render("new-pw-formHTML", fiber.Map{
				"Error": "Error: setting new password!",
			})
		}

		return c.Redirect("/")
	}
}
