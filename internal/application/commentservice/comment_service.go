package commentservice

import (
	"context"
	"filmPackager/internal/domain/comment"
	"filmPackager/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
)

type CommentService struct {
	CommentRepo comment.CommentRepository
	UserRepo    user.UserRepository
}

func NewCommentService(commentRepo comment.CommentRepository, userRepo user.UserRepository) *CommentService {
	return &CommentService{CommentRepo: commentRepo, UserRepo: userRepo}
}

type CommentResponse struct {
	ID        uuid.UUID
	DocID     uuid.UUID
	Text      string
	Author    user.User
	CreatedAt string
}

type GetDocCommentsResponse struct {
	Comments []CommentResponse
}

func (s *CommentService) GetDocComments(ctx context.Context, docID uuid.UUID) (*GetDocCommentsResponse, error) {
	rv := &GetDocCommentsResponse{}

	comments, err := s.CommentRepo.GetDocComments(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("error getting comments: %v", err)
	}

	uIDs := []uuid.UUID{}

	for _, c := range comments {
		rv.Comments = append(rv.Comments, CommentResponse{
			ID:        c.ID,
			DocID:     c.DocID,
			Text:      c.Content,
			CreatedAt: c.CreatedAt.Format("01-02-2006 -- 15:04"),
		})
		uIDs = append(uIDs, c.AuthorID)
	}

	users, err := s.UserRepo.GetUsersByIDs(ctx, uIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting users: %v", err)
	}

	userMap := map[uuid.UUID]user.User{}

	for _, u := range users {
		userMap[u.Id] = u
	}

	for i, c := range comments {
		rv.Comments[i].Author = userMap[c.AuthorID]
	}

	fmt.Println("rv: ", rv)
	return rv, nil
}

func (s *CommentService) CreateComment(ctx context.Context, text string, userID uuid.UUID, docID uuid.UUID) (*CommentResponse, error) {
	c := comment.CreateNewComment(docID, userID, text)
	rv := &CommentResponse{
		DocID:     docID,
		ID:        c.ID,
		Text:      c.Content,
		CreatedAt: c.CreatedAt.Format("01-02-2006 -- 15:04"),
	}

	err := s.CommentRepo.CreateDocComment(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("error creating comment: %v", err)
	}

	u, err := s.UserRepo.GetUserById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	rv.Author = *u

	return rv, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	err := s.CommentRepo.DeleteDocComment(ctx, commentID)
	if err != nil {
		return fmt.Errorf("error deleting comment: %v", err)
	}

	return nil
}
