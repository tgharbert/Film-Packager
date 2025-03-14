package comment

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID       uuid.UUID
	DocID    uuid.UUID
	Content  string
	AuthorID uuid.UUID
	// ADD A CREATED AT FIELD IN THE DB...
	CreatedAt time.Time
}

func CreateNewComment(docID, authorID uuid.UUID, comment string) *Comment {
	id := uuid.New()
	return &Comment{
		ID:        id,
		DocID:     docID,
		Content:   comment,
		AuthorID:  authorID,
		CreatedAt: time.Now(),
	}
}

func isByCurrUser(authorID, currUserID uuid.UUID) bool {
	return authorID == currUserID
}
