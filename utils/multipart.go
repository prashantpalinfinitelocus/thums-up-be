package utils

import (
	"bufio"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

const (
	// Large buffer size to handle big multipart uploads (1MB buffer)
	multipartBufferSize = 1024 * 1024
)

// ParseMultipartFormWithLargeBuffer parses a multipart form with a larger buffer
// to avoid "bufio: buffer full" errors with large files
func ParseMultipartFormWithLargeBuffer(r *http.Request, maxMemory int64) error {
	if r.MultipartForm != nil {
		return nil // Already parsed
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return http.ErrNotMultipart
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		return http.ErrNotMultipart
	}

	boundary, ok := params["boundary"]
	if !ok {
		return http.ErrMissingBoundary
	}

	// Create a buffered reader with a large buffer
	bufferedReader := bufio.NewReaderSize(r.Body, multipartBufferSize)
	
	// Create multipart reader with the buffered reader
	reader := multipart.NewReader(bufferedReader, boundary)

	// Parse the multipart form
	form, err := reader.ReadForm(maxMemory)
	if err != nil {
		return err
	}

	r.MultipartForm = form
	return nil
}

// GetFormFileWithLargeBuffer gets a file from multipart form with large buffer support
func GetFormFileWithLargeBuffer(r *http.Request, key string, maxMemory int64) (*multipart.FileHeader, error) {
	// Parse with large buffer if not already parsed
	if err := ParseMultipartFormWithLargeBuffer(r, maxMemory); err != nil && err != http.ErrNotMultipart {
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

// GetPostFormWithLargeBuffer gets form values with large buffer support
func GetPostFormWithLargeBuffer(r *http.Request, key string, maxMemory int64) (string, error) {
	// Parse with large buffer if not already parsed
	if err := ParseMultipartFormWithLargeBuffer(r, maxMemory); err != nil && err != http.ErrNotMultipart {
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

