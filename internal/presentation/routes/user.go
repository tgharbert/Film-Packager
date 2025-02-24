package routes

import (
	"filmPackager/internal/application/userservice"
	access "filmPackager/internal/auth"
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

		createdUser, err := svc.CreateUser(c.Context(), firstName, lastName, email, password, secondPassword)
		if err != nil {
			return c.Render("create-accountHTML", fiber.Map{
				"Error": err.Error(),
			})
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
		email := strings.TrimSpace(c.FormValue("email"))
		password := strings.TrimSpace(c.FormValue("password"))

		currentUser, err := svc.UserLogin(c.Context(), email, password)
		if err != nil {
			return c.Render("login-formHTML", fiber.Map{
				"Error": err.Error(),
			})
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
		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("password1"))
		pw2 := strings.TrimSpace(c.FormValue("password2"))

		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}
		tokenString = tokenString[len("Bearer "):]

		userInfo, err := access.GetUserNameFromToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
		}

		// verify that the pw is correct
		err = svc.VerifyOldPassword(c.Context(), userInfo.Id, pw1, pw2)
		if err != nil {
			return c.Render("reset-passwordHTML", fiber.Map{
				"Error": err.Error(),
			})
		}

		return c.Render("new-pw-formHTML", userInfo)
	}
}

func SetNewPassword(svc *userservice.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// send the passwords to the user service
		pw1 := strings.TrimSpace(c.FormValue("new-password1"))
		pw2 := strings.TrimSpace(c.FormValue("new-password2"))

		tokenString := c.Cookies("Authorization")
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		u, err := access.GetUserNameFromToken(tokenString)

		err = svc.SetNewPassword(c.Context(), u.Id, pw1, pw2)
		if err != nil {
			return c.Render("new-pw-formHTML", fiber.Map{
				"Error": err.Error(),
			})
		}

		return c.Redirect("/")
	}
}
