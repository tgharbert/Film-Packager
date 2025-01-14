package document

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *Document) error
	FindStagedByType(ctx context.Context, orgID uuid.UUID, fileType string) (*Document, error)
	Delete(ctx context.Context, doc *Document) error
	GetAllByOrgId(ctx context.Context, orgID uuid.UUID) ([]*Document, error)
	// GetKeysForDeleteAll(ctx context.Context, orgID int) ([]string, error)
	GetDocumentDetails(ctx context.Context, docID uuid.UUID) (*Document, error)
	FindStagedByOrganization(ctx context.Context, orgID uuid.UUID) ([]*Document, error)
	UpdateDocument(ctx context.Context, doc *Document) error
	GetAllLockedDocumentsByProjectID(ctx context.Context, orgID uuid.UUID) ([]*Document, error)
	DeleteAllLockedByProjectID(ctx context.Context, orgID uuid.UUID) error
	UpdateAllStagedToLocked(ctx context.Context, orgID uuid.UUID) error
	DeleteSelectedDocuments(ctx context.Context, dIDs []uuid.UUID) error
}

type S3Repository interface {
	UploadFile(ctx context.Context, doc *Document, fileBody interface{}) (string, error)
	DeleteFile(ctx context.Context, doc *Document) error
	DeleteAllOrgFiles(ctx context.Context, keys []string) error
	// yet to write
	DownloadFile(ctx context.Context, fileName string, ID uuid.UUID) (*s3.GetObjectOutput, error)
}
