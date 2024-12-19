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

type UploadDocumentResponse struct {
	ID   uuid.UUID
	Date string
}

func (s *DocumentService) UploadDocument(ctx context.Context, orgID, userID uuid.UUID, fileName, fileType string, fileBody interface{}) (map[string]UploadDocumentResponse, error) {
	// check if repos are nil
	if s.docRepo == nil || s.s3Repo == nil {
		return nil, fmt.Errorf("nil repository")
	}

	// create a return value
	rv := make(map[string]UploadDocumentResponse)

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
	oldDoc, err := s.docRepo.FindStagedByType(ctx, orgID, fileType)
	switch err {
	// if there is an existing document, update the values
	case nil:
		// delete the file from the s3 bucket
		err := s.s3Repo.DeleteFile(ctx, oldDoc)
		if err != nil {
			return nil, fmt.Errorf("error deleting file: %v", err)
		}

		// upload the file to s3
		_, err = s.s3Repo.UploadFile(ctx, d, fileBody)
		if err != nil {
			return nil, fmt.Errorf("error uploading file: %v", err)
		}

		// update the document in the PG database
		err = s.docRepo.UpdateDocument(ctx, d)
		if err != nil {
			fmt.Printf("error updating document", err)
			return nil, fmt.Errorf("error updating document: %v", err)
		}
	// if there is no existing document, save the new document
	case document.ErrDocumentNotFound:
		// upload the file to s3
		_, err := s.s3Repo.UploadFile(ctx, d, fileBody)
		if err != nil {
			return nil, fmt.Errorf("error uploading file: %v", err)
		}

		// save to the PG database
		err = s.docRepo.Save(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("error saving document: %v", err)
		}

	// otherwise return the error
	default:
		return nil, fmt.Errorf("error finding staged document: %v", err)
	}

	docs, err := s.docRepo.GetAllByOrgId(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("error getting all documents: %v", err)
	}

	for _, doc := range docs {
		// we only need to return the staged documents
		if doc.IsStaged() {
			docResp := &UploadDocumentResponse{
				ID:   doc.ID,
				Date: doc.Date.Format("01-02-2006"),
			}
			rv[doc.FileType] = *docResp
		}
	}

	// we return the map of staged documents
	return rv, nil
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

func (s *DocumentService) LockDocuments(ctx context.Context, pID uuid.UUID) error {
	// delete the previous locked documents
	lockedDocs, err := s.docRepo.GetAllLockedDocumentsByProjectID(ctx, pID)
	if err != nil {
		return fmt.Errorf("error getting locked documents: %v", err)
	}

	// delete the files from the s3 bucket
	lockedNames := []string{}
	for _, doc := range lockedDocs {
		lockedNames = append(lockedNames, doc.FileName)
	}

	// delete the files from the s3 bucket
	err = s.s3Repo.DeleteAllOrgFiles(ctx, lockedNames)
	if err != nil {
		return fmt.Errorf("error deleting files: %v", err)
	}

	// delete the documents from the PG database
	err = s.docRepo.DeleteAllLockedByProjectID(ctx, pID)
	if err != nil {
		return fmt.Errorf("error deleting documents: %v", err)
	}

	// go through all staged docs and update "staged" to "locked" in PG
	err = s.docRepo.UpdateAllStagedToLocked(ctx, pID)
	if err != nil {
		return fmt.Errorf("error updating staged to locked: %v", err)
	}

	// only returning an error bc it would need to do so much work, get docs-membmerships-p details, etc
	return nil
}
