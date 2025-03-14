package routes

import (
	"filmPackager/internal/application/commentservice"
	"filmPackager/internal/application/middleware/auth"
	"fmt"

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

		comments, err := svc.GetDocComments(c.Context(), docUUID)
		if err != nil {
			fmt.Println("error getting comments: ", err)
			return c.Status(fiber.StatusInternalServerError).SendString("error getting comments")
		}

		return c.Render("document-commentsHTML", fiber.Map{
			"Comments": comments,
			"DocID":    docId,
		})
	}
}

func AddDocComment(svc *commentservice.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Println("hit the add comment route")
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		fmt.Println("docUUID: ", docUUID)
		comment := c.FormValue("comment")
		fmt.Println("comment: ", comment)
		u := auth.GetUserFromContext(c)

		nc, err := svc.CreateComment(c.Context(), comment, u.Id, docUUID)
		if err != nil {
			fmt.Println("error creating comment: ", err)
			return c.Status(fiber.StatusInternalServerError).SendString("error adding comment")
		}

		fmt.Println("new comment created: ", nc)

		return c.Redirect("/get-document/" + docId + "/")
	}
}

func DeleteComment(svc *commentservice.CommentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentId := c.Params("comment_id")
		commentUUID, err := uuid.Parse(commentId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		err = svc.DeleteComment(c.Context(), commentUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting comment")
		}

		return c.Redirect("/get-document/" + c.Params("doc_id") + "/")
	}
}
