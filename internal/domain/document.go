package domain

import "time"

type Document struct {
	ID             int
	OrganizationID int
	UserID         int
	FileName       string
	FileType       string
	Status         string
	Date           time.Time
	Color          string
}

func (d *Document) IsStaged() bool {
	return d.Status == "staged"
}
