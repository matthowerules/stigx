package models

import (
	"errors"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "validation error for field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestFileError_Error(t *testing.T) {
	cause := errors.New("underlying error")
	err := FileError{
		Operation: "test_operation",
		FilePath:  "/test/path",
		Cause:     cause,
	}

	expected := "file test_operation failed for '/test/path': underlying error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestFileError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := FileError{
		Operation: "test_operation",
		FilePath:  "/test/path",
		Cause:     cause,
	}

	if !errors.Is(err, cause) {
		t.Error("Expected FileError to unwrap to cause")
	}
}

func TestDatabaseError_Error(t *testing.T) {
	cause := errors.New("underlying error")
	err := DatabaseError{
		Operation: "test_operation",
		Table:     "test_table",
		Cause:     cause,
	}

	expected := "database test_operation failed on table 'test_table': underlying error"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestNotFoundError_Error(t *testing.T) {
	err := NotFoundError{
		Resource: "TestResource",
		ID:       "test_id",
	}

	expected := "TestResource with ID 'test_id' not found"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrorHandler_IsValidationError(t *testing.T) {
	handler := NewErrorHandler()

	// Test with ValidationError
	validationErr := ValidationError{Field: "test", Message: "test"}
	if !handler.IsValidationError(validationErr) {
		t.Error("Expected IsValidationError to return true for ValidationError")
	}

	// Test with other error
	regularErr := errors.New("regular error")
	if handler.IsValidationError(regularErr) {
		t.Error("Expected IsValidationError to return false for regular error")
	}
}

func TestErrorHandler_IsFileError(t *testing.T) {
	handler := NewErrorHandler()

	// Test with FileError
	fileErr := FileError{Operation: "test", FilePath: "/test"}
	if !handler.IsFileError(fileErr) {
		t.Error("Expected IsFileError to return true for FileError")
	}

	// Test with other error
	regularErr := errors.New("regular error")
	if handler.IsFileError(regularErr) {
		t.Error("Expected IsFileError to return false for regular error")
	}
}

func TestErrorHandler_IsDatabaseError(t *testing.T) {
	handler := NewErrorHandler()

	// Test with DatabaseError
	dbErr := DatabaseError{Operation: "test", Table: "test"}
	if !handler.IsDatabaseError(dbErr) {
		t.Error("Expected IsDatabaseError to return true for DatabaseError")
	}

	// Test with other error
	regularErr := errors.New("regular error")
	if handler.IsDatabaseError(regularErr) {
		t.Error("Expected IsDatabaseError to return false for regular error")
	}
}

func TestErrorHandler_IsNotFoundError(t *testing.T) {
	handler := NewErrorHandler()

	// Test with NotFoundError
	notFoundErr := NotFoundError{Resource: "test", ID: "test"}
	if !handler.IsNotFoundError(notFoundErr) {
		t.Error("Expected IsNotFoundError to return true for NotFoundError")
	}

	// Test with other error
	regularErr := errors.New("regular error")
	if handler.IsNotFoundError(regularErr) {
		t.Error("Expected IsNotFoundError to return false for regular error")
	}
}

func TestErrorHandler_GetRootCause(t *testing.T) {
	handler := NewErrorHandler()

	// Test with wrapped error
	rootCause := errors.New("root cause")
	wrappedErr := FileError{
		Operation: "test",
		FilePath:  "/test",
		Cause:     rootCause,
	}

	result := handler.GetRootCause(wrappedErr)
	if result != rootCause {
		t.Error("Expected GetRootCause to return the root cause")
	}

	// Test with unwrapped error
	regularErr := errors.New("regular error")
	result = handler.GetRootCause(regularErr)
	if result != regularErr {
		t.Error("Expected GetRootCause to return the error itself when not wrapped")
	}
}

func TestErrorHandler_HandleError(t *testing.T) {
	handler := NewErrorHandler()

	// Test ValidationError
	validationErr := ValidationError{Field: "test_field", Message: "test message"}
	userMsg, _, isRecoverable := handler.HandleError(validationErr)

	expectedUserMsg := "Invalid input: test message"
	if userMsg != expectedUserMsg {
		t.Errorf("Expected user message '%s', got '%s'", expectedUserMsg, userMsg)
	}

	if !isRecoverable {
		t.Error("Expected ValidationError to be recoverable")
	}

	// Test FileError
	fileErr := FileError{Operation: "read", FilePath: "/test/path"}
	userMsg, _, isRecoverable = handler.HandleError(fileErr)

	expectedUserMsg = "File operation failed. Please check the file and try again."
	if userMsg != expectedUserMsg {
		t.Errorf("Expected user message '%s', got '%s'", expectedUserMsg, userMsg)
	}

	if !isRecoverable {
		t.Error("Expected FileError to be recoverable")
	}

	// Test unknown error
	unknownErr := errors.New("unknown error")
	userMsg, _, isRecoverable = handler.HandleError(unknownErr)

	expectedUserMsg = "An unexpected error occurred. Please try again."
	if userMsg != expectedUserMsg {
		t.Errorf("Expected user message '%s', got '%s'", expectedUserMsg, userMsg)
	}

	if isRecoverable {
		t.Error("Expected unknown error to not be recoverable")
	}
}
