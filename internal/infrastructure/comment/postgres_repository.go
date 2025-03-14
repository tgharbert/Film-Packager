package infrastructure

import (
	"context"
	"filmPackager/internal/domain/comment"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCommentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCommentRepository(db *pgxpool.Pool) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

func (r *PostgresCommentRepository) CreateDocComment(ctx context.Context, comment *comment.Comment) error {
	query := `INSERT INTO comments (id, doc_id, user_id, text, date) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query, comment.ID, comment.DocID, comment.AuthorID, comment.Content, comment.CreatedAt)

	return err
}

func (r *PostgresCommentRepository) GetDocComments(ctx context.Context, docID uuid.UUID) ([]comment.Comment, error) {
	query := `SELECT id, document_id, user_id, comment, created_at FROM doc_comments WHERE document_id = $1`

	rows, err := r.db.Query(ctx, query, docID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []comment.Comment

	for rows.Next() {
		var comment comment.Comment

		err = rows.Scan(&comment.ID, &comment.DocID, &comment.AuthorID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *PostgresCommentRepository) DeleteDocComments(ctx context.Context, docID uuid.UUID) error {
	query := `DELETE FROM comments WHERE doc_id = $1`

	_, err := r.db.Exec(ctx, query, docID)

	return err
}

func (r *PostgresCommentRepository) DeleteDocComment(ctx context.Context, commentID uuid.UUID) error {
	query := `DELETE FROM comments WHERE id = $1`

	_, err := r.db.Exec(ctx, query, commentID)

	return err
}

func (r *PostgresCommentRepository) GetDocComment(ctx context.Context, commentID uuid.UUID) (*comment.Comment, error) {
	query := `SELECT id, doc_id, user_id, text, date FROM comments WHERE id = $1`

	row := r.db.QueryRow(ctx, query, commentID)

	var comment comment.Comment

	err := row.Scan(&comment.ID, &comment.DocID, &comment.AuthorID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}
