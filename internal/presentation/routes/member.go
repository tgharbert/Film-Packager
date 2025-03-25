package routes

import (
	"filmPackager/internal/application/membershipservice"
	"filmPackager/internal/domain/project"
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

		m, err := svc.UpdateMemberRoles(c.Context(), projUUID, mUserUUID, role)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error updating member roles")
		}

		var availRoles []string

		allRoles := []string{"director", "producer", "writer", "cinematographer", "production_designer"}
		for _, role := range allRoles {
			if slices.Contains(m.Roles, role) {
				continue
			} else {
				availRoles = append(availRoles, role)
			}
		}

		return c.Render("member-detailsHTML", fiber.Map{"Member": m, "ProjectId": projectId, "Roles": availRoles})
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

func SearchMembersByName(svc *membershipservice.MembershipService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.FormValue("username")
		projectID := c.Params("id")

		// parse the project id
		projUUID, err := uuid.Parse(projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// search for new members
		users, err := svc.SearchForNewMembersByName(c.Context(), name, projUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to query users")
		}

		if len(users) == 0 {
			return c.Render("search-resultsHTML", fiber.Map{
				"Error":     fmt.Sprintf("No users found for '%v'", name),
				"ProjectID": projectID,
			})
		}

		return c.Render("search-resultsHTML", fiber.Map{
			"SearchedMembers": users,
			"ProjectID":       projectID,
		})
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

		invitedMembers, err := svc.InviteUserToProject(c.Context(), userUUID, projUUID)
		if err != nil {
			if err == project.ErrMemberAlreadyInvited {
				// return the proper html fragment
				return c.Render("project-list", fiber.Map{
					"Error": "You've already invited this user!",
				})
			}
			return c.Status(fiber.StatusInternalServerError).SendString("error inviting user to project")
		}

		return c.Render("form-search-membersHTML", fiber.Map{
			"Invited": invitedMembers,
		})
	}
}
