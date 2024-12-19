package infrastructure

import (
	"context"
	"filmPackager/internal/domain/document"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3DocumentRepository struct {
	client *s3.Client
	bucket string
}

func NewS3DocumentRepository(client *s3.Client, bucket string) *S3DocumentRepository {
	return &S3DocumentRepository{client: client, bucket: bucket}
}

func (r *S3DocumentRepository) UploadFile(ctx context.Context, doc *document.Document, file interface{}) (string, error) {
	key := doc.FileName

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
	key := doc.FileName

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *S3DocumentRepository) DeleteAllOrgFiles(ctx context.Context, keys []string) error {
	// this is boiler plate found here: https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html
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

func (r *S3DocumentRepository) DownloadFile(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	}
	resp, err := r.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download the file from the s3: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read s3 object data: %w", err)
	}
	return data, nil
}
