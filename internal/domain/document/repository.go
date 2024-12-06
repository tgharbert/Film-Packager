package document

import "context"

type DocumentRepository interface {
	Save(ctx context.Context, doc *Document) error
	FindStagedByType(ctx context.Context, orgID int, fileType string) (*Document, error)
	Delete(ctx context.Context, doc *Document) error
	GetAllByOrgId(ctx context.Context, orgID int) ([]*Document, error)
	// GetKeysForDeleteAll(ctx context.Context, orgID int) ([]string, error)
	GetDocumentDetails(ctx context.Context, docID int) (*Document, error)
	FindStagedByOrganization(ctx context.Context, orgID int) ([]*Document, error)
}

type S3Repository interface {
	UploadFile(ctx context.Context, doc *Document, fileBody interface{}) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
	DeleteAllOrgFiles(ctx context.Context, keys []string) error
	// yet to write
	//	DownloadFile(ctx context.Context, key string) error
}
