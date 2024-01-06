package apperr

import (
	"errors"
	"testing"
)

func TestErrorUnknown(t *testing.T) {
	err := NewError("create.user", ErrorUnknown, "error creating user", "123", errors.New("unknown database error"))

	expected := "operation: create.user, error_code: unknown_error, message: error creating user, request_id: 123, internal_error: unknown database error"
	if err.Error() != expected {
		t.Errorf("Expected %s, but got %s", expected, err.Error())
	}
}

func TestErrorValidation(t *testing.T) {
	err := NewError("create.user", ErrorValidation, "email is required", "456", nil)

	expected := "operation: create.user, error_code: validation_error, message: email is required, request_id: 456, internal_error: nil"
	if err.Error() != expected {
		t.Errorf("Expected %s, but got %s", expected, err.Error())
	}
}

func TestErrorNotFound(t *testing.T) {
	err := NewError("get.user", ErrorNotFound, "user not found", "101", errors.New("sql: no rows in result set"))

	expected := "operation: get.user, error_code: not_found_error, message: user not found, request_id: 101, internal_error: sql: no rows in result set"
	if err.Error() != expected {
		t.Errorf("Expected %s, but got %s", expected, err.Error())
	}
}

func TestErrorForbidden(t *testing.T) {
	err := NewError("update.user", ErrorForbidden, "access denied user is not admin", "202", nil)

	expected := "operation: update.user, error_code: forbidden_error, message: access denied user is not admin, request_id: 202, internal_error: nil"
	if err.Error() != expected {
		t.Errorf("Expected %s, but got %s", expected, err.Error())
	}
}
