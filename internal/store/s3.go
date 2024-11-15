package s3Conn

import (
	"fmt"
	"log"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetS3Session() *s3.S3 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	return s3.New(sess)
}

func WriteToS3(s3Client *s3.S3, bucket string, key string, f multipart.File) (error) {
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
		Body: f,
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}
	return nil
}

func DeleteS3Object (s3Client *s3.S3, bucket string, key string) error {
	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error deleting file from s3: %v", err)
	}
	return nil
}

func DeleteMultipleS3Objects(s3Client *s3.S3, bucket string, keys []string) error {
	objectsToDelete := make([]*s3.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objectsToDelete[i] = &s3.ObjectIdentifier{
			Key: aws.String(key),
		}
	}
	input := &s3.DeleteObjectsInput{
		Bucket : aws.String(bucket),
		Delete: &s3.Delete{
			Objects: objectsToDelete,
			Quiet: aws.Bool(true),
		},
	}
	output, err := s3Client.DeleteObjects(input)
	if err != nil {
		return fmt.Errorf("failed to delete objects: %w", err)
	}
	for _, deleted := range output.Deleted {
		log.Printf("Deleted object: %s", *deleted.Key)
	}
	return nil
}