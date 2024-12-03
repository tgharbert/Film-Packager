package document

import "time"

type Document struct {
	ID             int
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
func NewDocument(organizationID, userID int, fileName, fileType, status, color string, date *time.Time) *Document {
	return &Document{
		OrganizationID: organizationID,
		UserID:         userID,
		FileName:       fileName,
		FileType:       fileType,
		Status:         status,
		Date:           date,
		Color:          color,
	}
}
