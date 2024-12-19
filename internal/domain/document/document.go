package document

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	UserID         uuid.UUID
	FileName       string
	FileType       string
	Status         string
	Date           *time.Time
	Color          string
}

func (d *Document) IsStaged() bool {
	return d.Status == "staged"
}
