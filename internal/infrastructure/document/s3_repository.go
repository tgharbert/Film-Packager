package infrastructure

import (
	"context"
	"errors"
	"filmPackager/internal/domain/document"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3DocumentRepository struct {
	client *s3.Client
	bucket string
}

func NewS3DocumentRepository(client *s3.Client, bucket string) *S3DocumentRepository {
	return &S3DocumentRepository{client: client, bucket: bucket}
}

func (r *S3DocumentRepository) UploadFile(ctx context.Context, doc *document.Document, file interface{}) (string, error) {
	key := fmt.Sprintf("%s=%s", doc.FileName, doc.ID)

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		Body:   file.(multipart.File),
	})

	if err != nil {
		return "", err
	}

	return doc.FileName, nil
}

func (r *S3DocumentRepository) DeleteFile(ctx context.Context, doc *document.Document) error {
	key := fmt.Sprintf("%s=%s", doc.FileName, doc.ID)

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}

	return nil
}

// this is boiler plate found here: https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
func (r *S3DocumentRepository) DeleteAllOrgFiles(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	objectsToDelete := make([]types.ObjectIdentifier, len(keys))

	for i, key := range keys {
		objectsToDelete[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	input := s3.DeleteObjectsInput{
		Bucket: aws.String(r.bucket),
		Delete: &types.Delete{
			Objects: objectsToDelete,
			Quiet:   aws.Bool(true),
		},
	}

	_, err := r.client.DeleteObjects(ctx, &input)
	if err != nil {
		return err
	}

	return nil
}

// should this be the io.Reader type?
// DownloadFile gets an object from a bucket and stores it in a local file.
func (r *S3DocumentRepository) DownloadFile(ctx context.Context, fileName string, id uuid.UUID) (*os.File, error) {
	key := fmt.Sprintf("%s=%s", fileName, id)
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get object. No such key exists.\n")
			err = noKey
		} else {
			log.Printf("Couldn't get object. Here's why: %v\n", err)
		}
		return nil, err
	}
	defer result.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
		return nil, err
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		file.Close() // Ensure to close the file on error
		return nil, fmt.Errorf("Failed to download from S3: %v", err)
	}

	if _, err := file.Write(body); err != nil {
		file.Close() // Ensure to close the file on error
		return nil, fmt.Errorf("Failed to write to file: %v", err)
	}

	// Return the open file handle
	log.Printf("Downloaded file: %s, bytes: %d\n", fileName, len(body))
	return file, nil
}
