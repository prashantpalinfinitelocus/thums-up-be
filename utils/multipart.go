package utils

import (
	"bufio"
	"io"
	"mime/multipart"
	"net/http"
)

// ParseMultipartFormLargeFiles parses multipart form with large memory limit
// This wraps the request body in a large buffer to avoid "bufio: buffer full" errors
func ParseMultipartFormLargeFiles(r *http.Request, maxMemory int64) error {
	if r.MultipartForm != nil {
		return nil // Already parsed
	}

	// Wrap the body in a buffered reader with a very large buffer (64MB)
	// This prevents "bufio: buffer full" errors when parsing large multipart forms
	if r.Body != nil {
		bufferedBody := bufio.NewReaderSize(r.Body, 64*1024*1024) // 64MB buffer
		r.Body = io.NopCloser(bufferedBody)
	}

	// Use standard Go parsing with large memory limit
	// Files larger than maxMemory will be automatically spilled to disk
	return r.ParseMultipartForm(maxMemory)
}

// GetFormFileLargeFiles gets a file from multipart form with large file support
func GetFormFileLargeFiles(r *http.Request, key string, maxMemory int64) (*multipart.FileHeader, error) {
	// Parse if not already parsed
	if err := ParseMultipartFormLargeFiles(r, maxMemory); err != nil {
		return nil, err
	}

	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, http.ErrMissingFile
	}

	files := r.MultipartForm.File[key]
	if len(files) == 0 {
		return nil, http.ErrMissingFile
	}

	return files[0], nil
}

// GetPostFormLargeFiles gets form values with large file support
func GetPostFormLargeFiles(r *http.Request, key string, maxMemory int64) (string, error) {
	// Parse if not already parsed
	if err := ParseMultipartFormLargeFiles(r, maxMemory); err != nil {
		return "", err
	}

	if r.MultipartForm == nil || r.MultipartForm.Value == nil {
		return "", nil
	}

	values := r.MultipartForm.Value[key]
	if len(values) == 0 {
		return "", nil
	}

	return values[0], nil
}

