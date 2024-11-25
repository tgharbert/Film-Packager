package application

import (
	"context"
	"filmPackager/internal/domain"
)

type DocumentRepository interface {
	Save(ctx context.Context, doc *domain.Document) error
	FindStagedByType(ctx context.Context, orgID int, fileType string) (*domain.Document, error)
	Delete(ctx context.Context, doc *domain.Document) error
	GetKeysForDeleteAll(ctx context.Context, orgID int) ([]string, error)
}

type S3Repository interface {
	UploadFile(ctx context.Context, doc *domain.Document, fileBody interface{}) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
	DeleteAllOrgFiles(ctx context.Context, keys []string) error
	// yet to write
	DownloadFile(ctx context.Context, key string) error
}

type DocumentService struct {
	docRepo DocumentRepository
	s3Repo  S3Repository
}

func NewDocumentService(docRepo DocumentRepository, s3Repo S3Repository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo}
}

// this should return the data for the html frag with the date or whatever
func (s *DocumentService) UploadDocument(ctx context.Context, doc *domain.Document, fileBody interface{}) error {
	existingDoc, err := s.docRepo.FindStagedByType(ctx, doc.OrganizationID, doc.FileType)
	if err != nil {
		return err
	}
	if existingDoc != nil {
		err = s.s3Repo.DeleteFile(ctx, existingDoc.FileName)
		if err != nil {
			return err
		}
		err = s.docRepo.Delete(ctx, existingDoc)
		if err != nil {
			return err
		}
	}
	fileName, err := s.s3Repo.UploadFile(ctx, doc, fileBody)
	if err != nil {
		return err
	}
	doc.FileName = fileName
	doc.Status = "staged"
	return s.docRepo.Save(ctx, doc)
}
