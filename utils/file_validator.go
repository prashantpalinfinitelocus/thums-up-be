package utils

import (
	"fmt"
	"mime/multipart"
	"strings"
)

const (
	MaxFileSize = 100 * 1024 * 1024 // 100MB in bytes
)

var (
	AllowedAudioTypes = []string{
		"audio/mpeg",   // .mp3
		"audio/mp3",    // .mp3
		"audio/wav",    // .wav
		"audio/wave",   // .wav
		"audio/x-wav",  // .wav
		"audio/aac",    // .aac
		"audio/mp4",    // .m4a
		"audio/x-m4a",  // .m4a
		"audio/ogg",    // .ogg
		"audio/webm",   // .webm
		"audio/flac",   // .flac
		"audio/x-flac", // .flac
	}

	AllowedVideoTypes = []string{
		"video/mp4",       // .mp4
		"video/mpeg",      // .mpeg
		"video/quicktime", // .mov
		"video/x-msvideo", // .avi
		"video/webm",      // .webm
		"video/ogg",       // .ogv
		"video/3gpp",      // .3gp
		"video/3gpp2",     // .3g2
		"video/x-ms-wmv",  // .wmv
		"video/x-flv",     // .flv
	}

	AllowedAudioExtensions = []string{
		".mp3", ".wav", ".aac", ".m4a", ".ogg", ".webm", ".flac",
	}

	AllowedVideoExtensions = []string{
		".mp4", ".mpeg", ".mpg", ".mov", ".avi", ".webm", ".ogv", ".3gp", ".3g2", ".wmv", ".flv",
	}
)

type FileValidationError struct {
	Message string
}

func (e *FileValidationError) Error() string {
	return e.Message
}

func ValidateMediaFile(file *multipart.FileHeader) error {
	if file == nil {
		return nil // File is optional
	}

	// Check file size
	if file.Size > MaxFileSize {
		return &FileValidationError{
			Message: fmt.Sprintf("File size exceeds maximum allowed size of 100MB. Current size: %.2fMB", float64(file.Size)/(1024*1024)),
		}
	}

	// Check if file size is 0
	if file.Size == 0 {
		return &FileValidationError{
			Message: "File is empty",
		}
	}

	// Get content type
	contentType := file.Header.Get("Content-Type")

	// Get file extension
	filename := strings.ToLower(file.Filename)

	// Check if it's a valid audio or video file
	isValidAudio := isValidAudioFile(contentType, filename)
	isValidVideo := isValidVideoFile(contentType, filename)

	if !isValidAudio && !isValidVideo {
		return &FileValidationError{
			Message: "Invalid file type. Only audio and video files are allowed. Supported formats: mp3, wav, aac, m4a, ogg, flac (audio) and mp4, mov, avi, webm (video)",
		}
	}

	return nil
}

func isValidAudioFile(contentType, filename string) bool {
	// Check by content type
	for _, allowedType := range AllowedAudioTypes {
		if contentType == allowedType {
			return true
		}
	}

	// Check by extension as fallback
	for _, ext := range AllowedAudioExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	return false
}

func isValidVideoFile(contentType, filename string) bool {
	// Check by content type
	for _, allowedType := range AllowedVideoTypes {
		if contentType == allowedType {
			return true
		}
	}

	// Check by extension as fallback
	for _, ext := range AllowedVideoExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	return false
}

func GetMediaType(file *multipart.FileHeader) string {
	contentType := file.Header.Get("Content-Type")
	filename := strings.ToLower(file.Filename)

	if isValidAudioFile(contentType, filename) {
		return "audio"
	}

	if isValidVideoFile(contentType, filename) {
		return "video"
	}

	return "unknown"
}


