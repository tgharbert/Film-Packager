package commentservice

import (
	"context"
	"filmPackager/internal/domain/comment"
	"fmt"

	"github.com/google/uuid"
)

type CommentService struct {
	docRepo comment.CommentRepository
}

func (s *CommentService) GetDocComments(ctx context.Context, docID uuid.UUID) ([]comment.Comment, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}

	comments, err := s.docRepo.GetDocComments(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments: %v", err)
	}

	return comments, nil
}
