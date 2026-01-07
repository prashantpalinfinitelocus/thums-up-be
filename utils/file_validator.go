package utils

import (
	"fmt"
	"mime/multipart"
	"strings"
)

const (
	MaxAudioSize      = 100 * 1024 * 1024 // 100MB in bytes
	MaxVideoSize      = 10 * 1024 * 1024  // 10MB in bytes
	MaxImageSize      = 10 * 1024 * 1024  // 10MB in bytes
	ContentTypeHeader = "Content-Type"
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

	AllowedImageTypes = []string{
		"image/jpeg",               // .jpg, .jpeg
		"image/jpg",                // .jpg
		"image/png",                // .png
		"image/gif",                // .gif
		"image/webp",               // .webp
		"image/svg+xml",            // .svg
		"image/bmp",                // .bmp
		"image/x-icon",             // .ico
		"image/vnd.microsoft.icon", // .ico
	}

	AllowedAudioExtensions = []string{
		".mp3", ".wav", ".aac", ".m4a", ".ogg", ".webm", ".flac",
	}

	AllowedVideoExtensions = []string{
		".mp4", ".mpeg", ".mpg", ".mov", ".avi", ".webm", ".ogv", ".3gp", ".3g2", ".wmv", ".flv",
	}

	AllowedImageExtensions = []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico",
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

	// Check if file size is 0
	if file.Size == 0 {
		return &FileValidationError{
			Message: "File is empty",
		}
	}

	// Get content type and filename
	contentType := file.Header.Get(ContentTypeHeader)
	filename := strings.ToLower(file.Filename)

	// Determine file type and validate size accordingly
	isValidAudio := isValidAudioFile(contentType, filename)
	isValidVideo := isValidVideoFile(contentType, filename)
	isValidImage := isValidImageFile(contentType, filename)

	if !isValidAudio && !isValidVideo && !isValidImage {
		return &FileValidationError{
			Message: "Invalid file type. Only audio, video, and image files are allowed. Supported formats: mp3, wav, aac, m4a, ogg, flac (audio, max 100MB), mp4, mov, avi, webm (video, max 10MB), and jpg, jpeg, png, gif, webp (image, max 10MB)",
		}
	}

	// Check file size based on file type
	if isValidAudio && file.Size > MaxAudioSize {
		return &FileValidationError{
			Message: fmt.Sprintf("Audio file size exceeds maximum allowed size of 100MB. Current size: %.2fMB", float64(file.Size)/(1024*1024)),
		}
	}

	if isValidVideo && file.Size > MaxVideoSize {
		return &FileValidationError{
			Message: fmt.Sprintf("Video file size exceeds maximum allowed size of 10MB. Current size: %.2fMB", float64(file.Size)/(1024*1024)),
		}
	}

	if isValidImage && file.Size > MaxImageSize {
		return &FileValidationError{
			Message: fmt.Sprintf("Image file size exceeds maximum allowed size of 10MB. Current size: %.2fMB", float64(file.Size)/(1024*1024)),
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
	contentType := file.Header.Get(ContentTypeHeader)
	filename := strings.ToLower(file.Filename)

	if isValidAudioFile(contentType, filename) {
		return "audio"
	}

	if isValidVideoFile(contentType, filename) {
		return "video"
	}

	if isValidImageFile(contentType, filename) {
		return "image"
	}

	return "unknown"
}

func ValidateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return &FileValidationError{
			Message: "Image file is required",
		}
	}

	// Check file size
	if file.Size > MaxImageSize {
		return &FileValidationError{
			Message: fmt.Sprintf("Image file size exceeds maximum allowed size of 10MB. Current size: %.2fMB", float64(file.Size)/(1024*1024)),
		}
	}

	// Check if file size is 0
	if file.Size == 0 {
		return &FileValidationError{
			Message: "File is empty",
		}
	}

	// Get content type
	contentType := file.Header.Get(ContentTypeHeader)

	// Get file extension
	filename := strings.ToLower(file.Filename)

	// Check if it's a valid image file
	if !isValidImageFile(contentType, filename) {
		return &FileValidationError{
			Message: "Invalid file type. Only image files are allowed. Supported formats: jpg, jpeg, png, gif, webp, svg, bmp, ico",
		}
	}

	return nil
}

func isValidImageFile(contentType, filename string) bool {
	// Check by content type
	for _, allowedType := range AllowedImageTypes {
		if contentType == allowedType {
			return true
		}
	}

	// Check by extension as fallback
	for _, ext := range AllowedImageExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	return false
}
