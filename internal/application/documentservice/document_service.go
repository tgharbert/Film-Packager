package documentservice

// package application

import (
	"context"
	"filmPackager/internal/domain/document"
	"filmPackager/internal/domain/membership"
	"filmPackager/internal/domain/project"
	"filmPackager/internal/domain/user"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type DocumentService struct {
	docRepo    document.DocumentRepository
	s3Repo     document.S3Repository
	userRepo   user.UserRepository
	memberRepo membership.MembershipRepository
	projRepo   project.ProjectRepository
}

func NewDocumentService(docRepo document.DocumentRepository, s3Repo document.S3Repository, userRepo user.UserRepository, memberRepo membership.MembershipRepository, projRepo project.ProjectRepository) *DocumentService {
	return &DocumentService{docRepo: docRepo, s3Repo: s3Repo, userRepo: userRepo, memberRepo: memberRepo, projRepo: projRepo}
}

type UploadDocumentResponse struct {
	ID   uuid.UUID
	Date string
}

type DownloadDocumentResponse struct {
	DocStream *s3.GetObjectOutput
	FileName  string
}

type GetDocumentDetailsResponse struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	FileName     string
	UploaderName string
	UploadDate   string
	DocType      string
	Status       string
}

// map of access tiers and the file types they can access
var accessTiers = map[string][]string{
	"owner":               {"Script", "Logline", "Synopsis", "PitchDeck", "Schedule", "Budget", "Shotlist", "Lookbook"},
	"director":            {"Script", "Logline", "Synopsis", "PitchDeck", "Schedule", "Budget", "Shotlist", "Lookbook"},
	"producer":            {"Script", "Logline", "Synopsis", "PitchDeck", "Schedule", "Budget", "Shotlist", "Lookbook"},
	"writer":              {"Script", "Logline", "Synopsis", "PitchDeck", "Lookbook"},
	"cinematographer":     {"PitchDeck", "Budget", "Shotlist", "Lookbook"},
	"production_designer": {"PitchDeck", "Budget", "Lookbook"},
}

// standardize the file naming on upload - ProjectName-FileType-Date
func (s *DocumentService) UploadDocument(ctx context.Context, orgID, userID uuid.UUID, fileName, fileType string, fileBody interface{}) (map[string]UploadDocumentResponse, error) {
	m, err := s.memberRepo.GetMembership(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting membership: %v", err)
	}

	// checks the users highest role against their access tier
	if !slices.Contains(accessTiers[m.Roles[0]], fileType) {
		// return custom error for template
		return nil, document.ErrAccessDenied
	}

	// create a return value
	rv := make(map[string]UploadDocumentResponse)

	// create uniform file name
	project, err := s.projRepo.GetProjectByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("error getting project: %v", err)
	}

	fileName = fmt.Sprintf("%s_%s_%s", project.Name, fileType, time.Now().Format("01-02-2006"))

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

func (s *DocumentService) GetDocumentDetails(ctx context.Context, docID uuid.UUID) (*GetDocumentDetailsResponse, error) {
	if s.docRepo == nil {
		return nil, fmt.Errorf("nil repository")
	}

	doc, err := s.docRepo.GetDocumentDetails(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("error getting document details: %v", err)
	}

	uploader, err := s.userRepo.GetUserById(ctx, doc.UserID)
	if err != nil {
		return nil, fmt.Errorf("error getting uploader details: %v", err)
	}

	rv := &GetDocumentDetailsResponse{
		ID:           doc.ID,
		OrgID:        doc.OrganizationID,
		FileName:     doc.FileName,
		UploaderName: uploader.Name,
		UploadDate:   doc.Date.Format("01-02-2006, 15:04"),
		DocType:      doc.FileType,
		Status:       doc.Status,
	}

	return rv, nil
}

func (s *DocumentService) LockDocuments(ctx context.Context, pID uuid.UUID, uID uuid.UUID) error {
	m, err := s.memberRepo.GetMembership(ctx, pID, uID)
	if err != nil {
		return fmt.Errorf("error getting membership: %v", err)
	}

	// check if the user has the correct access - only owner, director, producer can lock
	if !slices.Contains([]string{"owner", "director", "producer"}, m.Roles[0]) {
		return document.ErrAccessDenied
	}

	// get all the locked documents
	lockedDocs, err := s.docRepo.GetAllLockedDocumentsByProjectID(ctx, pID)
	if err != nil {
		return fmt.Errorf("error getting locked documents: %v", err)
	}

	// get all the staged documents
	stagedDocs, err := s.docRepo.FindStagedByOrganization(ctx, pID)
	if err != nil {
		return fmt.Errorf("error getting staged documents: %v", err)
	}

	// I only want to delete the files that are both locked and staged
	// so I will create a map of the staged documents for simpler access
	stagedMap := make(map[string]*document.Document)
	for _, doc := range stagedDocs {
		stagedMap[doc.FileType] = doc
	}

	// create a list of the locked documents that are also staged
	// keys for the s3 bucket and IDs for the PG database
	keysToDelete := []string{}
	IDsToDelete := []uuid.UUID{}
	for _, doc := range lockedDocs {
		if _, ok := stagedMap[doc.FileType]; ok {
			// format the key for the s3 bucket
			key := fmt.Sprintf("%s=%s", doc.FileName, doc.ID)
			keysToDelete = append(keysToDelete, key)
			IDsToDelete = append(IDsToDelete, doc.ID)
		}
	}

	// delete the previous locked files from the s3 bucket
	err = s.s3Repo.DeleteAllOrgFiles(ctx, keysToDelete)
	if err != nil {
		return fmt.Errorf("error deleting files: %v", err)
	}

	// delete the locked documents from the PG database only when there is a replacement available
	err = s.docRepo.DeleteSelectedDocuments(ctx, IDsToDelete)
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

// need to document and further understand
func (s *DocumentService) DownloadDocument(ctx context.Context, docID uuid.UUID) (DownloadDocumentResponse, error) {
	rv := DownloadDocumentResponse{}

	if s.s3Repo == nil {
		return rv, fmt.Errorf("nil repository")
	}

	// get the document details
	doc, err := s.docRepo.GetDocumentDetails(ctx, docID)
	if err != nil {
		return rv, fmt.Errorf("error getting document details: %v", err)
	}

	// download the file from the s3 bucket
	stream, err := s.s3Repo.DownloadFile(ctx, doc.FileName, doc.ID)
	if err != nil {
		return rv, fmt.Errorf("error downloading file: %v", err)
	}

	rv.DocStream = stream
	rv.FileName = doc.FileName
	return rv, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, docID uuid.UUID) (uuid.UUID, error) {
	pID := uuid.UUID{}

	// get the doc from the PG database
	doc, err := s.docRepo.GetDocumentDetails(ctx, docID)
	if err != nil {
		return pID, fmt.Errorf("error getting document details: %v", err)
	}

	// delete the file from the s3 bucket
	err = s.s3Repo.DeleteFile(ctx, doc)
	if err != nil {
		return pID, fmt.Errorf("error deleting file: %v", err)
	}

	// delete the document from the PG database
	err = s.docRepo.Delete(ctx, doc)
	if err != nil {
		return pID, fmt.Errorf("error deleting document: %v", err)
	}

	// return the project ID to redirect to the project page
	return doc.OrganizationID, nil
}
