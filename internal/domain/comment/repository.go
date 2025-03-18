package comment

import (
	"context"

	"github.com/google/uuid"
)

type CommentRepository interface {
	GetDocComments(ctx context.Context, docID uuid.UUID) ([]Comment, error)
	CreateDocComment(ctx context.Context, comment *Comment) error
	DeleteDocComments(ctx context.Context, docID uuid.UUID) error
	DeleteDocComment(ctx context.Context, commentID uuid.UUID) error
	GetDocComment(ctx context.Context, commentID uuid.UUID) (*Comment, error)
}
