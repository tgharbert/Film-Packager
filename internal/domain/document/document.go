package document

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID             uuid.UUID
	OrganizationID int
	UserID         int
	FileName       string
	FileType       string
	Status         string
	Date           *time.Time
	Color          string
}

func (d *Document) IsStaged() bool {
	return d.Status == "staged"
}

// create new document
func NewDocument(organizationID, userID int, fileName, fileType string) *Document {
	now := time.Now()
	id := uuid.New()

	return &Document{
		ID:             id,
		OrganizationID: organizationID,
		UserID:         userID,
		FileName:       fileName,
		FileType:       fileType,
		Status:         "staged",
		Date:           &now,
		Color:          "black",
	}
}
