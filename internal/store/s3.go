package s3

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GetS3Client(ctx context.Context) *s3.Client {
	// Load the default configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-2"))
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}
	// Create the S3 client
	return s3.NewFromConfig(cfg)
}

func WriteToS3(ctx context.Context, s3Client *s3.Client, bucket string, key string, f multipart.File) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}
	return nil
}

func DeleteS3Object(ctx context.Context, s3Client *s3.Client, bucket string, key string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error deleting file from s3: %v", err)
	}
	return nil
}

// func DeleteMultipleS3Objects(ctx context.Context, s3Client *s3.Client, bucket string, keys []string) error {
// 	objectsToDelete := make([]*s3.ObjectIdentifier, len(keys))
// 	for i, key := range keys {
// 		objectsToDelete[i] = &s3.ObjectIdentifier{
// 			Key: aws.String(key),
// 		}
// 	}
// 	input := &s3.DeleteObjectsInput{
// 		Bucket: aws.String(bucket),
// 		Delete: &s3.Delete{
// 			Objects: objectsToDelete,
// 			Quiet:   aws.Bool(true),
// 		},
// 	}
// 	output, err := s3Client.DeleteObjects(input)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete objects: %w", err)
// 	}
// 	for _, deleted := range output.Deleted {
// 		log.Printf("Deleted object: %s", *deleted.Key)
// 	}
// 	return nil
// }
