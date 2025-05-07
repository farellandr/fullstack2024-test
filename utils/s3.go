package utils

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type S3Service struct {
	S3       *s3.S3
	Uploader *s3manager.Uploader
	Bucket   string
}

func InitS3() *S3Service {
	awsRegion := getEnv("AWS_REGION", "ap-southeast-1a")
	awsAccessKey := getEnv("AWS_ACCESS_KEY_ID", "")
	awsSecretKey := getEnv("AWS_SECRET_ACCESS_KEY", "")
	awsBucket := getEnv("AWS_S3_BUCKET", "client-logos")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})

	if err != nil {
		log.Printf("Warning: Failed to create AWS session: %v", err)
		return nil
	}

	return &S3Service{
		S3:       s3.New(sess),
		Uploader: s3manager.NewUploader(sess),
		Bucket:   awsBucket,
	}
}

func (s *S3Service) UploadFile(file *multipart.FileHeader) (string, error) {
	openedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer openedFile.Close()

	buffer := make([]byte, file.Size)
	if _, err = openedFile.Read(buffer); err != nil {
		return "", err
	}

	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	_, err = s.Uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(buffer),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.Bucket, fileName), nil
}
