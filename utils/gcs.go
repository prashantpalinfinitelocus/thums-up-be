package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

type GCSService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (string, string, error)
	DeleteFile(ctx context.Context, fileURL string) error
	UploadFileFromReader(ctx context.Context, file io.ReadCloser, folder string) (string, string, error)
	UploadFileFromBytes(ctx context.Context, data []byte, path string, contentType string) (string, string, error)
	GetFileSignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
	ReadFileContent(ctx context.Context, objectPath string) (string, error)
	GetPublicURL(objectPath string) string
}

type gcsService struct {
	bucketName string
	client     *storage.Client
	projectID  string
}

func NewGCSService(bucketName, projectID string) (GCSService, error) {
	ctx := context.Background()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &gcsService{
		bucketName: bucketName,
		client:     client,
		projectID:  projectID,
	}, nil
}

func NewGCSServiceWithCredentials(bucketName, projectID, credentialsPath string) (GCSService, error) {
	ctx := context.Background()

	var client *storage.Client
	var err error

	if credentialsPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsPath))
	} else {
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &gcsService{
		bucketName: bucketName,
		client:     client,
		projectID:  projectID,
	}, nil
}

func (s *gcsService) UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", path, uuid.New().String(), ext)

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	writer := s.client.Bucket(s.bucketName).Object(filename).NewWriter(ctx)
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"original-filename": file.Filename,
	}
	writer.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(writer, src); err != nil {
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, filename)
	return url, filename, nil
}

func (s *gcsService) UploadFileFromReader(ctx context.Context, file io.ReadCloser, path string) (string, string, error) {
	defer file.Close()

	filename := fmt.Sprintf("%s/%s", path, uuid.New().String())

	writer := s.client.Bucket(s.bucketName).Object(filename).NewWriter(ctx)
	writer.ContentType = "application/octet-stream"
	writer.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(writer, file); err != nil {
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, filename)
	return url, filename, nil
}

func (s *gcsService) UploadFileFromBytes(ctx context.Context, data []byte, path string, contentType string) (string, string, error) {
	writer := s.client.Bucket(s.bucketName).Object(path).NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=86400"

	if _, err := writer.Write(data); err != nil {
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, path)
	return url, path, nil
}

func (s *gcsService) DeleteFile(ctx context.Context, fileURL string) error {
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.bucketName)
	if !strings.HasPrefix(fileURL, prefix) {
		return fmt.Errorf("invalid GCS URL format")
	}
	key := strings.TrimPrefix(fileURL, prefix)

	object := s.client.Bucket(s.bucketName).Object(key)
	if err := object.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete GCS object: %w", err)
	}

	return nil
}

func (s *gcsService) GetFileSignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	}
	url, err := s.client.Bucket(s.bucketName).SignedURL(objectPath, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

func (s *gcsService) ReadFileContent(ctx context.Context, objectPath string) (string, error) {
	if strings.HasPrefix(objectPath, "https://storage.googleapis.com/") {
		prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", s.bucketName)
		objectPath = strings.TrimPrefix(objectPath, prefix)
	}

	object := s.client.Bucket(s.bucketName).Object(objectPath)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create reader for GCS object: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read GCS object content: %w", err)
	}

	return string(content), nil
}

func (s *gcsService) GetPublicURL(objectPath string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectPath)
}
