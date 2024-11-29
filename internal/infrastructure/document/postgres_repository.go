package infrastructure

import (
	"context"
	"database/sql"
	"filmPackager/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDocumentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresDocumentRepository(db *pgxpool.Pool) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{db: db}
}

func (r *PostgresDocumentRepository) Save(ctx context.Context, doc *domain.Document) error {
	query := `INSERT INTO documents (organization_id, user_id, file_name, file_type, date, color, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, doc.OrganizationID, doc.UserID, doc.FileName, doc.FileType, doc.Date, doc.Color, doc.Status)
	return err
}

func (r *PostgresDocumentRepository) FindStagedByType(ctx context.Context, orgID int, fileType string) (*domain.Document, error) {
	checkStagedQuery := `SELECT id, organization_id, user_id, file_name, file_type, status, date, color FROM documents WHERE organization_id = $1 AND status = 'staged' AND file_type = $2`
	row := r.db.QueryRow(ctx, checkStagedQuery, orgID, fileType)
	var doc domain.Document
	err := row.Scan(&doc.ID, &doc.OrganizationID, &doc.UserID, &doc.FileName, &doc.FileType, &doc.Status, &doc.Date, &doc.Color)
	if err != sql.ErrNoRows {
		return nil, nil
	}
	return &doc, nil
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, doc *domain.Document) error {
	deleteQuery := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.Exec(ctx, deleteQuery, doc.ID)
	return err
}

func (r *PostgresDocumentRepository) GetKeysForDeleteAll(ctx, orgID int) (*[]string, error) {
	getKeysQuery := `SELECT file_name FROM documents where organization_id = $1`
	var keys []string
	rows, err := r.db.Query(context.Background(), getKeysQuery, orgID)
	if err != nil {
		return &keys, fmt.Errorf("error retrieving address from db: %v", err)
	}
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return &keys, fmt.Errorf("error scanning rows: %v", err)
		}
		keys = append(keys, key)
	}
	if rows.Err() != nil {
		return &keys, rows.Err()
	}
	return &keys, nil
}
