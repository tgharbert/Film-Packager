package routes

import (
	"filmPackager/internal/application/authservice"
	"filmPackager/internal/application/middleware/auth"
	"filmPackager/internal/application/userservice"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetLoginPage(svc *authservice.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// need to check if cookie is valid, if not render login
		tokenString := c.Cookies("filmpackager")
		if tokenString == "" {
			return c.Render("login-form", nil)
		}

		tokenString = tokenString[len("Bearer "):]

		err := svc.VerifyToken(tokenString)
		if err != nil {
			return c.Render("login-form", nil)
		}
		return c.Redirect("/")
	}
}

func PostCreateAccount(svc *authservice.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		firstName := strings.Trim(c.FormValue("firstName"), " ")
		lastName := strings.Trim(c.FormValue("lastName"), " ")
		email := strings.Trim(c.FormValue("email"), " ")
		password := strings.Trim(c.FormValue("password"), " ")
		secondPassword := strings.Trim(c.FormValue("secondPassword"), " ")

		tokenString, err := svc.CreateNewUser(c.Context(), firstName, lastName, email, password, secondPassword)
		if err != nil {
			return c.Render("create-accountHTML", fiber.Map{
				"Error": err.Error(),
			})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "filmpackager",
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

func LoginUserHandler(svc *authservice.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		email := strings.TrimSpace(c.FormValue("email"))
		password := strings.TrimSpace(c.FormValue("password"))

		// create login token
		tokenString, err := svc.CreateLoginToken(c.Context(), email, password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Error generating JWT")
		}

		c.Cookie(&fiber.Cookie{
			Name:     "filmpackager",
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
			Name:  "filmpackager",
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
		tokenString := c.Cookies("filmpackager")
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		return c.Render("reset-passwordHTML", nil)
	}
}

func VerifyOldPassword(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("password1"))
		pw2 := strings.TrimSpace(c.FormValue("password2"))

		tokenString := c.Cookies("filmpackager")
		if tokenString == "" {
			return c.Redirect("/login/")
		}
		tokenString = tokenString[len("Bearer "):]

		u := auth.GetUserFromContext(c)

		// verify that the pw is correct
		err := svc.VerifyOldPassword(c.Context(), u.Id, pw1, pw2)
		if err != nil {
			return c.Render("reset-passwordHTML", fiber.Map{
				"Error": err.Error(),
			})
		}

		return c.Render("new-pw-formHTML", u)
	}
}

func SetNewPassword(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the user from the context
		u := auth.GetUserFromContext(c)

		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("new-password1"))
		pw2 := strings.TrimSpace(c.FormValue("new-password2"))

		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		err := svc.SetNewPassword(c.Context(), u.Id, pw1, pw2)
		if err != nil {
			return c.Render("new-pw-formHTML", fiber.Map{
				"Error": err.Error(),
			})
		}

		return c.Redirect("/")
	}
}
