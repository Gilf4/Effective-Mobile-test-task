package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) String() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

func NewBadRequest(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: message,
		Err:     err,
	}
}

func NewNotFound(message string, err error) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: message,
		Err:     err,
	}
}

func NewInternal(err error) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
		Err:     err,
	}
}
