package routes

import (
	"filmPackager/internal/application/middleware/auth"
	"filmPackager/internal/application/projectservice"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetHomePage(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("filmpackager")
		if c.Get("HX-Request") == "true" {
			c.Set("HX-Redirect", "/") // Redirect to homepage or desired URL
			return nil
		}
		if tokenString == "" {
			return c.Redirect("/login/")
		}

		tokenString = tokenString[len("Bearer "):]

		u := auth.GetUserFromContext(c)

		rv, err := svc.GetUsersProjects(c.Context(), u)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}

		return c.Render("index", *rv)
	}
}

func GetProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// get the user from the cookie
		u := auth.GetUserFromContext(c)

		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		p, err := svc.GetProjectDetails(c.Context(), projUUID, u.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error retrieving project data")
		}

		return c.Render("project-page", *p)
	}
}

func ClickDeleteProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		p, err := svc.GetProject(c.Context(), projUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting project data")
		}

		return c.Render("projectDeleteCancelHTML", p)
	}
}

func CancelDeleteProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := auth.GetUserFromContext(c)

		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// get all the project info for this one project
		rv, err := svc.GetProjectOverview(c.Context(), projUUID, u.Id)

		return c.Render("project-list-item", fiber.Map{"ID": projUUID, "Roles": rv.Roles, "Name": rv.Name})
	}
}

func DeleteProject(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := auth.GetUserFromContext(c)

		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			fmt.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.DeleteProject(c.Context(), projUUID, u)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting project")
		}

		rv, err = svc.GetUsersProjects(c.Context(), u)
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

		u := auth.GetUserFromContext(c)

		project, err := svc.CreateNewProject(c.Context(), projectName, u.Id)
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

		u := auth.GetUserFromContext(c)

		err = svc.JoinProject(c.Context(), projUUID, u.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error joining project")
		}

		rv, err := svc.GetUsersProjects(c.Context(), u)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Error retrieving orgs")
		}

		return c.Render("selectOrgHTML", *rv)
	}
}

func GetUpdateNameForm(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectId := c.Params("project_id")
		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		p, err := svc.GetProject(c.Context(), projUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting project data")
		}

		return c.Render("edit-projectHTML", p)
	}
}

func UpdateProjectName(svc *projectservice.ProjectService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectName := c.FormValue("project-name")
		projectId := c.Params("project_id")

		projUUID, err := uuid.Parse(projectId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		p, err := svc.UpdateProjectName(c.Context(), projUUID, projectName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error updating project name")
		}

		return c.Render("edit-projectHTML", p)
	}
}
