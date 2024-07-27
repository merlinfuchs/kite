package handler

import (
	"fmt"
	"net/http"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Data    interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func ErrNotFound(code, message string) *Error {
	return &Error{
		Status:  http.StatusNotFound,
		Code:    code,
		Message: message,
	}
}

func ErrBadRequest(code, message string) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Code:    code,
		Message: message,
	}
}

func ErrForbidden(code, message string) *Error {
	return &Error{
		Status:  http.StatusForbidden,
		Code:    code,
		Message: message,
	}

}

func ErrUnauthorized(code, message string) *Error {
	return &Error{
		Status:  http.StatusUnauthorized,
		Code:    code,
		Message: message,
	}
}

func ErrInternal(message string) *Error {
	return &Error{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: message,
	}
}
