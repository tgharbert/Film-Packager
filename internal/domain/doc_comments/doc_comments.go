package doc_comment

import (
	"time"

	"github.com/google/uuid"
)

type DocComment struct {
	ID       uuid.UUID
	DocID    uuid.UUID
	Comment  string
	AuthorID uuid.UUID
	// ADD A CREATED AT FIELD IN THE DB...
	CreatedAt time.Time
}

func CreateNewDocComment(docID, authorID uuid.UUID, comment string) *DocComment {
	id := uuid.New()
	return &DocComment{
		ID:        id,
		DocID:     docID,
		Comment:   comment,
		AuthorID:  authorID,
		CreatedAt: time.Now(),
	}
}
