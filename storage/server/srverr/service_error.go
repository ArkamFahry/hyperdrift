package srverr

import (
	"errors"
	"fmt"
)

type ErrorCode error

var NotFoundError ErrorCode = errors.New("not found error")
var ConflictError ErrorCode = errors.New("conflict error")
var InvalidInputError ErrorCode = errors.New("invalid input error")
var BadRequestError ErrorCode = errors.New("bad request error")
var ForbiddenError ErrorCode = errors.New("forbidden error")
var UnknownError ErrorCode = errors.New("unknown error")

type ServiceError struct {
	ErrorCode     ErrorCode `json:"error_code"`
	Message       string    `json:"message"`
	Operation     string    `json:"operation"`
	RequestId     string    `json:"request_id"`
	InternalError error     `json:"internal_error"`
}

func NewServiceError(errorCode ErrorCode, message string, operation string, requestId string, internalError error) ServiceError {
	return ServiceError{
		ErrorCode:     errorCode,
		Message:       message,
		Operation:     operation,
		RequestId:     requestId,
		InternalError: internalError,
	}
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("error_code: %s, message: %s, operation: %s, request_id: %s, internal_error: %s", e.ErrorCode, e.Message, e.Operation, e.RequestId, e.InternalError)
}
