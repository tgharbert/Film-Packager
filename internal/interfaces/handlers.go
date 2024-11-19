package interfaces

import (
	"filmPackager/internal/application"
	"filmPackager/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func UploadDocumentHandler(svc *application.DocumentService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		orgID := c.FormValue("organization_id")
		fileType := c.FormValue("file_type")
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("File is required")
		}
		f, _ := file.Open()
		defer f.Close()
		orgIDInt, err := strconv.Atoi(orgID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("error parsing Id from request")
		}
		doc := &domain.Document{
			OrganizationID: orgIDInt,
			FileName:       file.Filename,
			FileType:       fileType,
		}
		err = svc.UploadDocument(c.Context(), doc, f)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		// return the HTML fragment/template
		return c.JSON(doc)
	}
}
