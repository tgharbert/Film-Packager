package routes

import (
	"filmPackager/internal/application/documentservice"
	"filmPackager/internal/application/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetDocCommentSection(svc *commentService.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		comments, err := svc.GetDocComments(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting comments")
		}

		return c.Render("comment-sectionHTML", fiber.Map{
			"Comments": comments,
		})
	}
}

func AddDocComment(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		comment := c.FormValue("comment")
		u := auth.GetUserFromContext(c)

		err = svc.AddComment(c.Context(), docUUID, u.Id, comment)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error adding comment")
		}

		return c.Redirect("/get-document/" + docId + "/")
	}
}
