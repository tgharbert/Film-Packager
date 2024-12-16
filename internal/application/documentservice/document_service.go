package documentservice

// package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/user"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DocumentService struct {
	docRepo  document.DocumentRepository
	s3Repo   document.S3Repository
	userRepo user.UserRepository
}

func NewDocumentService(docRepo document.DocumentRepository, s3Repo document.S3Repository, userRepo user.UserRepository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo, userRepo: userRepo}
}

func (s *DocumentService) UploadDocument(ctx context.Context, orgID, userID uuid.UUID, fileName, fileType string, fileBody interface{}) ([]document.Document, error) {
	// check if repos are nil
	if s.docRepo == nil || s.s3Repo == nil {
		return nil, fmt.Errorf("nil repository")
	}

	// create a new document object
	now := time.Now()
	d := &document.Document{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		FileName:       fileName,
		FileType:       fileType,
		Date:           &now,
		Status:         "staged",
		Color:          "black",
	}

	// check if there is a document with the same type for the org
	_, err := s.docRepo.FindStagedByType(ctx, orgID, fileType)

	switch err {
	// if there is an existing document, update the values
	case nil:
		err := s.docRepo.UpdateDocument(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("error updating document: %v", err)
		}

		// get all the documents for the project
		docs, err := s.docRepo.GetAllByOrgId(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("error getting all documents: %v", err)
		}
		fmt.Println("docs: ", docs)
		return []document.Document{*d}, nil

	// if there is no existing document, save the new document
	case document.ErrDocumentNotFound:
		break

	// otherwise return the error
	default:
		return nil, fmt.Errorf("error finding staged document: %v", err)
	}

	// save the document
	err = s.docRepo.Save(ctx, d)
	if err != nil {
		fmt.Println("error saving document: ", err)
		return nil, fmt.Errorf("error saving document: %v", err)
	}

	// should this return all the documents??
	return []document.Document{*d}, nil
}

func (s *DocumentService) GetDocumentDetails(ctx context.Context, docID uuid.UUID) (*document.Document, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	return s.docRepo.GetDocumentDetails(ctx, docID)
}

func (s *DocumentService) GetUploaderDetails(ctx context.Context, userId uuid.UUID) (*user.User, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}
	return s.userRepo.GetUserById(ctx, userId)
}
