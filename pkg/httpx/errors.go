package httpx

import (
	"net/http"

	"github.com/aussiebroadwan/taboo/sdk"
)

// Common error codes.
const (
	CodeNotFound   = "NOT_FOUND"
	CodeBadRequest = "BAD_REQUEST"
	CodeInternal   = "INTERNAL_ERROR"
)

// APIError represents an API error with a code and HTTP status.
type APIError struct {
	Code    string
	Message string
	Status  int
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return e.Message
}

// ErrNotFound creates a not found error.
func ErrNotFound(message string) *APIError {
	return &APIError{
		Code:    CodeNotFound,
		Message: message,
		Status:  http.StatusNotFound,
	}
}

// ErrBadRequest creates a bad request error.
func ErrBadRequest(message string) *APIError {
	return &APIError{
		Code:    CodeBadRequest,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// ErrInternal creates an internal server error.
func ErrInternal(message string) *APIError {
	return &APIError{
		Code:    CodeInternal,
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

// WriteError writes an APIError as a JSON response.
func WriteError(w http.ResponseWriter, err *APIError) error {
	return JSON(w, err.Status, sdk.ErrorResponse{
		Error: sdk.ErrorDetail{
			Code:    err.Code,
			Message: err.Message,
		},
	})
}
