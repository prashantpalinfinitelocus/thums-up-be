package utils

import (
	"bufio"
	"io"
	"mime/multipart"
	"net/http"
)

func ParseMultipartFormLargeFiles(r *http.Request, maxMemory int64) error {
	if r.MultipartForm != nil {
		return nil
	}

	if r.Body != nil {
		bufferedBody := bufio.NewReaderSize(r.Body, 150*1024*1024)
		r.Body = io.NopCloser(bufferedBody)
	}

	return r.ParseMultipartForm(maxMemory)
}

func GetFormFileLargeFiles(r *http.Request, key string, maxMemory int64) (*multipart.FileHeader, error) {
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

func GetPostFormLargeFiles(r *http.Request, key string, maxMemory int64) (string, error) {
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
