package routes

import (
	"filmPackager/internal/application/commentservice"
	"filmPackager/internal/application/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetDocCommentSection(svc *commentservice.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.GetDocComments(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting comments")
		}

		return c.Render("document-commentsHTML", fiber.Map{
			"Comments": rv.Comments,
			"DocID":    docId,
		})
	}
}

func AddDocComment(svc *commentservice.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		comment := c.FormValue("comment")
		u := auth.GetUserFromContext(c)

		nc, err := svc.CreateComment(c.Context(), comment, u.Id, docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error adding comment")
		}

		return c.Render("doc-commentHTML", nc)
	}
}

func DeleteComment(svc *commentservice.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentId := c.Params("comment_id")
		commentUUID, err := uuid.Parse(commentId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.DeleteComment(c.Context(), commentUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting comment")
		}

		// figure out what to return here...
		return c.Render("doc-comments-listHTML", fiber.Map{
			"Comments": rv.Comments,
			"DocID":    rv.DocID,
		})
	}
}
