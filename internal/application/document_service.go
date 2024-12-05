package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"fmt"
)

type DocumentService struct {
	docRepo document.DocumentRepository
	s3Repo  document.S3Repository
}

func NewDocumentService(docRepo document.DocumentRepository, s3Repo document.S3Repository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo}
}

func (s *DocumentService) UploadDocument(ctx context.Context, doc *document.Document, fileBody interface{}) (string, error) {
	// check if repos are nil
	if s.docRepo == nil || s.s3Repo == nil {
		return "", fmt.Errorf("nil repository")
	}
	if doc == nil {
		return "", fmt.Errorf("nil document")
	}
	existingDoc, err := s.docRepo.FindStagedByType(ctx, doc.OrganizationID, doc.FileType)
	if err != nil {
		fmt.Println("error finding existing doc", err)
		return "", err
	}
	if existingDoc != nil {
		// delete the file from s3
		err = s.s3Repo.DeleteFile(ctx, existingDoc.FileName)
		if err != nil {
			return "", err
		}
		// delete the doc info from the pg db
		err = s.docRepo.Delete(ctx, existingDoc)
		if err != nil {
			return "", err
		}
	}
	fileName, err := s.s3Repo.UploadFile(ctx, doc, fileBody)
	if err != nil {
		return "", err
	}
	// do some work with the Time and send it back?
	doc.FileName = fileName
	// this should be done on the retrieval??
	timeStr := doc.Date.Format("2006-01-02 15:04")
	return timeStr, s.docRepo.Save(ctx, doc)
}

func (s *DocumentService) GetDocumentDetails(ctx context.Context, docID int) (*document.Document, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	return s.docRepo.GetDocumentDetails(ctx, docID)
}
