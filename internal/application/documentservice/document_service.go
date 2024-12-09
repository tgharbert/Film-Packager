package documentservice

// package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/user"
	"fmt"
)

type DocumentService struct {
	docRepo  document.DocumentRepository
	s3Repo   document.S3Repository
	userRepo user.UserRepository
}

func NewDocumentService(docRepo document.DocumentRepository, s3Repo document.S3Repository, userRepo user.UserRepository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo, userRepo: userRepo}
}

func (s *DocumentService) UploadDocument(ctx context.Context, doc *document.Document, fileBody interface{}) ([]document.Document, error) {
	// check if repos are nil
	if s.docRepo == nil || s.s3Repo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	existingDoc, err := s.docRepo.FindStagedByType(ctx, doc.OrganizationID, doc.FileType)
	if err != nil {
		fmt.Println("error finding existing doc", err)
		return nil, err
	}
	if existingDoc != nil {
		err = s.s3Repo.DeleteFile(ctx, existingDoc.FileName)
		if err != nil {
			return nil, err
		}
		err = s.docRepo.Delete(ctx, existingDoc)
		if err != nil {
			return nil, err
		}
	}
	fileName, err := s.s3Repo.UploadFile(ctx, doc, fileBody)
	if err != nil {
		return nil, err
	}
	doc.FileName = fileName
	err = s.docRepo.Save(ctx, doc)
	if err != nil {
		return nil, err
	}
	documents, err := s.docRepo.FindStagedByOrganization(ctx, doc.OrganizationID)
	if err != nil {
		return nil, err
	}
	var staged []document.Document
	for _, doc := range documents {
		staged = append(staged, *doc)
	}
	return staged, nil
}

func (s *DocumentService) GetDocumentDetails(ctx context.Context, docID int) (*document.Document, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	return s.docRepo.GetDocumentDetails(ctx, docID)
}

func (s *DocumentService) GetUploaderDetails(ctx context.Context, userId int) (*user.User, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	return s.userRepo.GetUserById(ctx, userId)
}
