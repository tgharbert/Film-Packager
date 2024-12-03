package application

import (
	"context"
	"filmPackager/internal/domain/document"
)

type DocumentService struct {
	docRepo document.DocumentRepository
	s3Repo  document.S3Repository
}

func NewDocumentService(docRepo document.DocumentRepository, s3Repo document.S3Repository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo}
}

// this should return the data for the html frag with the date or whatever
func (s *DocumentService) UploadDocument(ctx context.Context, doc *document.Document, fileBody interface{}) error {
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
