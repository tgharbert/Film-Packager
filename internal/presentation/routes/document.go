package routes

import (
	"filmPackager/internal/application/documentservice"
	access "filmPackager/internal/auth"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func LockStagedDocs(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		pIDString := c.Params("project_id")
		pID, err := uuid.Parse(pIDString)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		err = svc.LockDocuments(c.Context(), pID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error locking documents")
		}

		return c.Redirect("/get-project/" + pIDString + "/")
	}
}

func DownloadDocument(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		rv, err := svc.DownloadDocument(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error downloading document")
		}
		defer rv.DocStream.Body.Close()

		// need the file name...
		attachment := fmt.Sprintf("attachment; filename=%s", rv.FileName)

		// Set the appropriate headers
		c.Set("Content-Type", "application/octet-stream") // Adjust Content-Type as needed
		c.Set("Content-Disposition", attachment)

		// Stream the body to the client
		if _, err := io.Copy(c, rv.DocStream.Body); err != nil {
			fmt.Println("error copying file to response", err)
			return c.Status(fiber.StatusInternalServerError).SendString("error copying file to response")
		}

		return nil
	}
}

func UploadDocumentHandler(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.Params("project_id")
		fileType := c.FormValue("file-type")
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("File is required")
		}

		userInfo, err := access.GetUserDataFromCookie(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting user info from cookie")
		}

		f, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error opening file")
		}
		defer f.Close()

		orgUUID, err := uuid.Parse(orgID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// returns a map of staged documents
		documents, err := svc.UploadDocument(c.Context(), orgUUID, userInfo.Id, file.Filename, fileType, f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		return c.Render("staged-listHTML", fiber.Map{
			"Staged": documents,
		})
	}
}

func GetDocDetails(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		// combine these into one call -- Rett
		doc, err := svc.GetDocumentDetails(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting document details")
		}

		// need to get the user info as well
		uploadingUser, err := svc.GetUploaderDetails(c.Context(), doc.UserID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting uploader details")
		}

		return c.Render("document-detailsHTML", fiber.Map{
			"Document": *doc,
			"Uploader": *uploadingUser,
		})
	}
}
