package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"filmPackager/internal/domain/document"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDocumentRepository struct {
	db *pgxpool.Pool
}

func NewPostgresDocumentRepository(db *pgxpool.Pool) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{db: db}
}

// GetAllByOrgId returns all documents for a given organization
func (r *PostgresDocumentRepository) GetAllByOrgId(ctx context.Context, orgID uuid.UUID) ([]*document.Document, error) {
	query := `SELECT id, organization_id, user_id, file_name, file_type, status, date, color FROM documents WHERE organization_id = $1`

	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving documents from db: %v", err)
	}

	defer rows.Close()

	var docs []*document.Document

	for rows.Next() {
		var doc document.Document

		err = rows.Scan(&doc.ID, &doc.OrganizationID, &doc.UserID, &doc.FileName, &doc.FileType, &doc.Status, &doc.Date, &doc.Color)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}

		docs = append(docs, &doc)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return docs, nil
}

func (r *PostgresDocumentRepository) Save(ctx context.Context, doc *document.Document) error {
	query := `INSERT INTO documents (id, organization_id, user_id, file_name, file_type, date, color, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(ctx, query, doc.ID, doc.OrganizationID, doc.UserID, doc.FileName, doc.FileType, doc.Date, doc.Color, doc.Status)

	return err
}

func (r *PostgresDocumentRepository) UpdateDocument(ctx context.Context, doc *document.Document) error {
	// update the document with the same org_id and file_type and status = 'staged'
	query := `UPDATE documents SET id = $1, user_id = $2, file_name = $3, file_type = $4, status = $5, date = $6, color = $7 WHERE organization_id = $8 AND file_type = $4 AND status = 'staged'`

	_, err := r.db.Exec(ctx, query, doc.ID, doc.UserID, doc.FileName, doc.FileType, doc.Status, doc.Date, doc.Color, doc.OrganizationID)
	if err != nil {
		return fmt.Errorf("error updating document: %v", err)
	}

	return err
}

func (r *PostgresDocumentRepository) GetDocumentDetails(ctx context.Context, docID uuid.UUID) (*document.Document, error) {
	query := `SELECT id, organization_id, user_id, file_name, file_type, status, date, color FROM documents WHERE id = $1`

	row := r.db.QueryRow(ctx, query, docID)

	var doc document.Document

	err := row.Scan(&doc.ID, &doc.OrganizationID, &doc.UserID, &doc.FileName, &doc.FileType, &doc.Status, &doc.Date, &doc.Color)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %v", err)
	}

	return &doc, nil
}

func (r *PostgresDocumentRepository) FindStagedByType(ctx context.Context, orgID uuid.UUID, fileType string) (*document.Document, error) {
	checkStagedQuery := `SELECT id, organization_id, user_id, file_name, file_type, status, date, color FROM documents WHERE organization_id = $1 AND status = 'staged' AND file_type = $2`

	row := r.db.QueryRow(ctx, checkStagedQuery, orgID, fileType)

	var doc document.Document

	err := row.Scan(&doc.ID, &doc.OrganizationID, &doc.UserID, &doc.FileName, &doc.FileType, &doc.Status, &doc.Date, &doc.Color)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, document.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("error scanning row: %v", err)
	}
	return &doc, nil
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, doc *document.Document) error {
	deleteQuery := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.Exec(ctx, deleteQuery, doc.ID)
	return err
}

func (r *PostgresDocumentRepository) GetKeysForDeleteAll(ctx, orgID uuid.UUID) (*[]string, error) {
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

func (r *PostgresDocumentRepository) FindStagedByOrganization(ctx context.Context, orgID uuid.UUID) ([]*document.Document, error) {
	query := `SELECT id, organization_id, user_id, file_name, file_type, status, date, color FROM documents WHERE organization_id = $1 AND status = 'staged'`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving documents from db: %v", err)
	}
	defer rows.Close()
	var docs []*document.Document
	for rows.Next() {
		var doc document.Document
		err = rows.Scan(&doc.ID, &doc.OrganizationID, &doc.UserID, &doc.FileName, &doc.FileType, &doc.Status, &doc.Date, &doc.Color)
		if err != nil {
			return nil, fmt.Errorf("error scanning rows: %v", err)
		}
		docs = append(docs, &doc)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return docs, nil
}
