package routes

import (
	"filmPackager/internal/application/documentservice"
	"filmPackager/internal/application/middleware/auth"
	"filmPackager/internal/domain/document"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func LockStagedDocs(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		u := auth.GetUserFromContext(c)

		pIDString := c.Params("project_id")
		pID, err := uuid.Parse(pIDString)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		err = svc.LockDocuments(c.Context(), pID, u.Id)
		if err != nil {
			if err == document.ErrAccessDenied {
				// alert the user that they don't have permission to lock the documents
				return c.Status(fiber.StatusOK).SendString("Access denied.")
			}
			//return c.Status(fiber.StatusInternalServerError).SendString("error locking documents")
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

		u := auth.GetUserFromContext(c)

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
		documents, err := svc.UploadDocument(c.Context(), orgUUID, u.Id, file.Filename, fileType, f)
		if err != nil {
			// if the user doesn't have permission to upload the document type
			if err == document.ErrAccessDenied {
				// UPDATE: this should send the template for the access error
				return c.Status(fiber.StatusUnauthorized).SendString("error uploading document")
			}
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

		rv, err := svc.GetDocumentDetails(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error getting document details")
		}

		return c.Render("document-detailsHTML", *rv)
	}
}

func DeleteDocument(svc *documentservice.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		docId := c.Params("doc_id")
		docUUID, err := uuid.Parse(docId)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}

		pID, err := svc.DeleteDocument(c.Context(), docUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error deleting document")
		}

		url := fmt.Sprintf("/get-project/%s/", pID)
		return c.Redirect(url)
	}
}
