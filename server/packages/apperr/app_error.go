package apperr

import (
	"errors"
	"strings"
)

type ErrorCode error

var (
	ErrorUnknown    ErrorCode = errors.New("unknown_error")
	ErrorValidation ErrorCode = errors.New("validation_error")
	ErrorNotFound   ErrorCode = errors.New("not_found_error")
	ErrorForbidden  ErrorCode = errors.New("forbidden_error")
)

type Error struct {
	Operation     string    `json:"operation"`
	ErrorCode     ErrorCode `json:"error_code"`
	Message       string    `json:"message"`
	RequestId     string    `json:"request_id"`
	InternalError error     `json:"internal_error"`
}

func (e *Error) Error() string {
	var errorBuilder strings.Builder

	errorBuilder.WriteString("operation: ")
	if e.Operation != "" {
		errorBuilder.WriteString(e.Operation)
	} else {
		errorBuilder.WriteString("nil")
	}

	errorBuilder.WriteString(", error_code: ")
	if e.ErrorCode != nil {
		errorBuilder.WriteString(e.ErrorCode.Error())
	} else {
		errorBuilder.WriteString(ErrorUnknown.Error())
	}

	errorBuilder.WriteString(", message: ")
	if e.Message != "" {
		errorBuilder.WriteString(e.Message)
	} else {
		errorBuilder.WriteString("nil")
	}

	errorBuilder.WriteString(", request_id: ")
	if e.RequestId != "" {
		errorBuilder.WriteString(e.RequestId)
	} else {
		errorBuilder.WriteString("nil")
	}

	errorBuilder.WriteString(", internal_error: ")
	if e.InternalError != nil {
		errorBuilder.WriteString(e.InternalError.Error())
	} else {
		errorBuilder.WriteString("nil")
	}

	return errorBuilder.String()
}

func NewError(operation string, errorCode ErrorCode, message string, requestId string, internalError error) *Error {
	return &Error{
		Operation:     operation,
		ErrorCode:     errorCode,
		Message:       message,
		RequestId:     requestId,
		InternalError: internalError,
	}
}
