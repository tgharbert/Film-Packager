package document

import "errors"

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrAccessDenied     = errors.New("member access blocked")
)
