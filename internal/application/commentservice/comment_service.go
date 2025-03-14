package commentservice

import (
	"context"
	"filmPackager/internal/domain/comment"
	"fmt"

	"github.com/google/uuid"
)

type CommentService struct {
	CommentRepo comment.CommentRepository
}

func NewCommentService(commentRepo comment.CommentRepository) *CommentService {
	return &CommentService{CommentRepo: commentRepo}
}

func (s *CommentService) GetDocComments(ctx context.Context, docID uuid.UUID) ([]comment.Comment, error) {
	if s.CommentRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}

	comments, err := s.CommentRepo.GetDocComments(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments: %v", err)
	}

	return comments, nil
}

func (s *CommentService) CreateComment(ctx context.Context, text string, userID uuid.UUID, docID uuid.UUID) (*comment.Comment, error) {
	c := comment.CreateNewComment(docID, userID, text)

	err := s.CommentRepo.CreateDocComment(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("error creating comment: %v", err)
	}

	return c, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	err := s.CommentRepo.DeleteDocComment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("error deleting comment: %v", err)
	}

	return nil
}
