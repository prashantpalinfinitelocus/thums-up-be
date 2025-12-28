package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
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

	service := &gcsService{
		bucketName: bucketName,
		client:     client,
		projectID:  projectID,
	}

	// Ensure bucket exists, create if it doesn't
	if err := service.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return service, nil
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

	service := &gcsService{
		bucketName: bucketName,
		client:     client,
		projectID:  projectID,
	}

	// Ensure bucket exists, create if it doesn't
	if err := service.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return service, nil
}

func (s *gcsService) UploadFile(ctx context.Context, file *multipart.FileHeader, path string) (string, string, error) {
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Get and normalize file extension from filename
	ext := strings.ToLower(filepath.Ext(file.Filename))
	originalFilename := file.Filename

	// Log for debugging
	log.Debugf("Uploading file: original=%s, extension=%s", originalFilename, ext)

	if ext == "" {
		// If no extension, try to detect from Content-Type
		contentTypeFromHeader := file.Header.Get("Content-Type")
		log.Debugf("No extension found, trying Content-Type: %s", contentTypeFromHeader)
		if contentTypeFromHeader != "" {
			ext = getExtensionFromContentType(contentTypeFromHeader)
			log.Debugf("Detected extension from Content-Type: %s", ext)
		}
	}

	// Generate unique filename with extension
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	// Full path for GCS storage
	gcsObjectPath := fmt.Sprintf("%s/%s", path, uniqueFilename)

	// Get Content-Type - prioritize header, then extension-based detection
	contentType := file.Header.Get("Content-Type")
	log.Debugf("Content-Type from header: %s", contentType)

	// Always validate and correct Content-Type based on extension to prevent mismatches
	if contentType == "" || contentType == "application/octet-stream" || !isValidContentTypeForExtension(contentType, ext) {
		// Try to detect MIME type from file extension
		detectedType := mime.TypeByExtension(ext)
		if detectedType != "" {
			contentType = detectedType
			log.Debugf("Detected Content-Type from mime package: %s", contentType)
		} else {
			// Fallback to extension-based detection
			contentType = getContentTypeFromExtension(ext)
			log.Debugf("Detected Content-Type from extension: %s", contentType)
		}
	}

	log.Debugf("Final: gcsPath=%s, filename=%s, extension=%s, contentType=%s", gcsObjectPath, uniqueFilename, ext, contentType)

	writer := s.client.Bucket(s.bucketName).Object(gcsObjectPath).NewWriter(ctx)
	writer.ContentType = contentType
	writer.Metadata = map[string]string{
		"original-filename": file.Filename,
	}
	writer.CacheControl = "public, max-age=86400"
	// Note: ACL is not set here because bucket uses Uniform Bucket-Level Access
	// Public access is controlled via bucket-level IAM policy

	if _, err := io.Copy(writer, src); err != nil {
		return "", "", fmt.Errorf("failed to write to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, gcsObjectPath)
	// Return URL and only the filename (not full path) for database storage
	return url, uniqueFilename, nil
}

func (s *gcsService) UploadFileFromReader(ctx context.Context, file io.ReadCloser, path string) (string, string, error) {
	defer file.Close()

	filename := fmt.Sprintf("%s/%s", path, uuid.New().String())

	writer := s.client.Bucket(s.bucketName).Object(filename).NewWriter(ctx)
	writer.ContentType = "application/octet-stream"
	writer.CacheControl = "public, max-age=86400"
	// Note: ACL is not set here because bucket uses Uniform Bucket-Level Access
	// Public access is controlled via bucket-level IAM policy

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
	// Note: ACL is not set here because bucket uses Uniform Bucket-Level Access
	// Public access is controlled via bucket-level IAM policy

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
	// objectPath should be the full path (e.g., "avatars/{userID}/{filename}")
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectPath)
}

// getContentTypeFromExtension returns the MIME type for common file extensions
// that might not be detected by mime.TypeByExtension
func getContentTypeFromExtension(ext string) string {
	ext = strings.ToLower(ext)

	// Image formats (check first for avatar/image uploads)
	imageTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".bmp":  "image/bmp",
		".ico":  "image/x-icon",
	}

	// Video formats
	videoTypes := map[string]string{
		".mov":  "video/quicktime",
		".mp4":  "video/mp4",
		".mpeg": "video/mpeg",
		".mpg":  "video/mpeg",
		".avi":  "video/x-msvideo",
		".webm": "video/webm",
		".ogv":  "video/ogg",
		".3gp":  "video/3gpp",
		".3g2":  "video/3gpp2",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
	}

	// Audio formats
	audioTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",
		".ogg":  "audio/ogg",
		".webm": "audio/webm",
		".flac": "audio/flac",
	}

	// Check image types first (prioritize for avatar uploads)
	if mimeType, ok := imageTypes[ext]; ok {
		return mimeType
	}

	// Check video types
	if mimeType, ok := videoTypes[ext]; ok {
		return mimeType
	}

	// Check audio types
	if mimeType, ok := audioTypes[ext]; ok {
		return mimeType
	}

	// Default fallback
	return "application/octet-stream"
}

// isValidContentTypeForExtension checks if the Content-Type matches the file extension
func isValidContentTypeForExtension(contentType, ext string) bool {
	ext = strings.ToLower(ext)
	contentType = strings.ToLower(contentType)

	// Map of extensions to valid content types
	validTypes := map[string][]string{
		".jpg":  {"image/jpeg", "image/jpg"},
		".jpeg": {"image/jpeg", "image/jpg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".webp": {"image/webp"},
		".svg":  {"image/svg+xml"},
		".bmp":  {"image/bmp"},
		".ico":  {"image/x-icon", "image/vnd.microsoft.icon"},
		".mov":  {"video/quicktime"},
		".mp4":  {"video/mp4"},
		".mp3":  {"audio/mpeg"},
		".wav":  {"audio/wav", "audio/wave", "audio/x-wav"},
	}

	if types, ok := validTypes[ext]; ok {
		for _, validType := range types {
			if contentType == validType {
				return true
			}
		}
		return false
	}

	// If extension not in our map, allow any non-octet-stream type
	return contentType != "application/octet-stream"
}

// getExtensionFromContentType returns the file extension based on Content-Type
func getExtensionFromContentType(contentType string) string {
	contentType = strings.ToLower(contentType)

	contentTypeToExt := map[string]string{
		"image/jpeg":      ".jpg",
		"image/jpg":       ".jpg",
		"image/png":       ".png",
		"image/gif":       ".gif",
		"image/webp":      ".webp",
		"image/svg+xml":   ".svg",
		"image/bmp":       ".bmp",
		"image/x-icon":    ".ico",
		"video/quicktime": ".mov",
		"video/mp4":       ".mp4",
		"video/mpeg":      ".mpg",
		"video/x-msvideo": ".avi",
		"video/webm":      ".webm",
		"audio/mpeg":      ".mp3",
		"audio/wav":       ".wav",
		"audio/aac":       ".aac",
		"audio/mp4":       ".m4a",
	}

	if ext, ok := contentTypeToExt[contentType]; ok {
		return ext
	}

	return ""
}

// ensureBucketExists checks if the bucket exists and creates it if it doesn't
func (s *gcsService) ensureBucketExists(ctx context.Context) error {
	if s.bucketName == "" {
		return fmt.Errorf("bucket name is required")
	}

	if s.projectID == "" {
		return fmt.Errorf("project ID is required to create bucket")
	}

	bucket := s.client.Bucket(s.bucketName)

	// Check if bucket exists
	_, err := bucket.Attrs(ctx)
	if err == nil {
		// Bucket exists, nothing to do
		log.Debugf("Bucket %s already exists", s.bucketName)
		return nil
	}

	// Check if error is a permission error (403) - service account might not have storage.buckets.get
	errStr := err.Error()
	isPermissionError := strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "Permission") ||
		strings.Contains(errStr, "permission denied") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "storage.buckets.get")

	if isPermissionError {
		// Service account doesn't have permission to check bucket existence
		// This is okay - we'll assume the bucket exists and let the first upload confirm
		log.Warnf(
			"Cannot verify bucket %s existence: service account lacks 'storage.buckets.get' permission. "+
				"Assuming bucket exists. First upload will confirm bucket status.",
			s.bucketName,
		)
		return nil
	}

	// Check if error is "bucket not found" - this can be ErrBucketNotExist or a wrapped error
	isNotFound := errors.Is(err, storage.ErrBucketNotExist)
	if !isNotFound {
		// Check if it's a 404 error (bucket not found)
		if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") || strings.Contains(errStr, "NotFound") {
			isNotFound = true
		}
	}

	if !isNotFound {
		// Unknown error when checking bucket existence
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	// Bucket doesn't exist, try to create it
	log.Infof("Bucket %s does not exist, attempting to create it in project %s", s.bucketName, s.projectID)
	createErr := bucket.Create(ctx, s.projectID, &storage.BucketAttrs{
		Location: "US", // Default location, can be made configurable if needed
	})

	if createErr == nil {
		log.Infof("Successfully created bucket %s", s.bucketName)
		return nil
	}

	// Check if the error is a permission error (403)
	createErrStr := createErr.Error()
	isCreatePermissionError := strings.Contains(createErrStr, "403") ||
		strings.Contains(createErrStr, "Permission") ||
		strings.Contains(createErrStr, "permission denied") ||
		strings.Contains(createErrStr, "forbidden")

	if isCreatePermissionError {
		// Service account doesn't have permission to create buckets
		// This might be okay if:
		// 1. The bucket exists but we got a different error when checking
		// 2. The bucket will be auto-created by Firebase/App Engine on first use
		// 3. The bucket will be created manually by an admin
		log.Warnf(
			"Cannot create bucket %s: service account lacks 'storage.buckets.create' permission. "+
				"Assuming bucket exists or will be created automatically/manually. "+
				"If uploads fail, ensure the bucket exists in project %s",
			s.bucketName, s.projectID,
		)

		// Try one more time to verify bucket exists (maybe it was just created or we had a transient error)
		time.Sleep(500 * time.Millisecond)
		_, verifyErr := bucket.Attrs(ctx)
		if verifyErr == nil {
			log.Infof("Bucket %s verified and accessible", s.bucketName)
			return nil
		}

		// If we can't verify, log a warning but continue anyway
		// The actual upload will fail with a clear error if the bucket truly doesn't exist
		log.Warnf("Could not verify bucket %s exists, but continuing. First upload will confirm bucket status.", s.bucketName)
		return nil
	}

	// Other errors (network, invalid config, etc.) should fail
	return fmt.Errorf("failed to create bucket %s: %w", s.bucketName, createErr)
}
