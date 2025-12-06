package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Err:        err,
	}
}

func NewUnauthorizedError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
		Err:        err,
	}
}

func NewNotFoundError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
		Err:        err,
	}
}

func NewInternalServerError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
		Err:        err,
	}
}

func NewConflictError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Message:    message,
		Err:        err,
	}
}

func NewTooManyRequestsError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusTooManyRequests,
		Message:    message,
		Err:        err,
	}
}

