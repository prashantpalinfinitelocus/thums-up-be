package errors

import (
	"errors"
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

func (e *AppError) Unwrap() error {
	return e.Err
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

func NewForbiddenError(message string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Message:    message,
		Err:        err,
	}
}

// Sentinel errors for type checking
var (
	ErrStateNotFound       = errors.New("state not found")
	ErrCityNotFound        = errors.New("city not found")
	ErrPincodeNotFound     = errors.New("pincode not found")
	ErrNotDeliverable      = errors.New("pincode not deliverable")
	ErrUserNotFound        = errors.New("user not found")
	ErrAddressNotFound     = errors.New("address not found")
	ErrAddressUnauthorized = errors.New("address does not belong to user")
)
