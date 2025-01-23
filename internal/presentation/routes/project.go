package routes

import (
	"filmPackager/internal/application/projectservice"
	access "filmPackager/internal/auth"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the user from the cookie
		u, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}

		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		p, err := svc.GetProjectDetails(c.Context(), projUUID, u.Id)
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("error retrieving project data")
		}

		return c.Render("project-page", *p)
	}
}

func DeleteProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}

		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.DeleteProject(c.Context(), projUUID, userInfo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting project")
		}

		rv, err = svc.GetUsersProjects(c.Context(), userInfo)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}

		return c.Render("selectOrgHTML", *rv)
	}
}

func CreateProject(svc *projectservice.ProjectService) fiber.Handler {
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
