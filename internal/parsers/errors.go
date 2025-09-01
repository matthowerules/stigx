package parsers

import (
	"fmt"
	"strings"
)

// ErrorType represents the category of parsing error
type ErrorType int

const (
	ErrorTypeUnknown ErrorType = iota
	ErrorTypeFileNotFound
	ErrorTypeFilePermission
	ErrorTypeInvalidFormat
	ErrorTypeValidationFailed
	ErrorTypeCorruptedData
	ErrorTypeMissingField
	ErrorTypeInvalidField
	ErrorTypeUnsupportedVersion
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeFileNotFound:
		return "FileNotFound"
	case ErrorTypeFilePermission:
		return "FilePermission"
	case ErrorTypeInvalidFormat:
		return "InvalidFormat"
	case ErrorTypeValidationFailed:
		return "ValidationFailed"
	case ErrorTypeCorruptedData:
		return "CorruptedData"
	case ErrorTypeMissingField:
		return "MissingField"
	case ErrorTypeInvalidField:
		return "InvalidField"
	case ErrorTypeUnsupportedVersion:
		return "UnsupportedVersion"
	default:
		return "Unknown"
	}
}

// ParseError represents a structured parsing error
type ParseError struct {
	Type     ErrorType
	Message  string
	Field    string
	Line     int
	Column   int
	Context  string
	Original error
}

// Error implements the error interface
func (pe *ParseError) Error() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("[%s]", pe.Type.String()))
	
	if pe.Field != "" {
		parts = append(parts, fmt.Sprintf("Field '%s':", pe.Field))
	}
	
	parts = append(parts, pe.Message)
	
	if pe.Line > 0 {
		if pe.Column > 0 {
			parts = append(parts, fmt.Sprintf("(line %d, column %d)", pe.Line, pe.Column))
		} else {
			parts = append(parts, fmt.Sprintf("(line %d)", pe.Line))
		}
	}
	
	if pe.Context != "" {
		parts = append(parts, fmt.Sprintf("Context: %s", pe.Context))
	}
	
	return strings.Join(parts, " ")
}

// Unwrap returns the underlying error
func (pe *ParseError) Unwrap() error {
	return pe.Original
}

// NewParseError creates a new ParseError
func NewParseError(errorType ErrorType, message string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Message: message,
	}
}

// NewParseErrorWithField creates a new ParseError with field information
func NewParseErrorWithField(errorType ErrorType, message, field string) *ParseError {
	return &ParseError{
		Type:    errorType,
		Message: message,
		Field:   field,
	}
}

// NewParseErrorWithLocation creates a new ParseError with location information
func NewParseErrorWithLocation(errorType ErrorType, message string, line, column int) *ParseError {
	return &ParseError{
		Type:    errorType,
		Message: message,
		Line:    line,
		Column:  column,
	}
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid   bool
	Errors  []*ParseError
	Warning []string
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(err *ParseError) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, err)
}

// AddWarning adds a validation warning
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warning = append(vr.Warning, message)
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// HasWarnings returns true if there are validation warnings
func (vr *ValidationResult) HasWarnings() bool {
	return len(vr.Warning) > 0
}

// ErrorSummary returns a summary of all errors
func (vr *ValidationResult) ErrorSummary() string {
	if !vr.HasErrors() {
		return "No errors"
	}
	
	var summary []string
	errorCounts := make(map[ErrorType]int)
	
	for _, err := range vr.Errors {
		errorCounts[err.Type]++
	}
	
	for errorType, count := range errorCounts {
		if count == 1 {
			summary = append(summary, fmt.Sprintf("1 %s error", errorType.String()))
		} else {
			summary = append(summary, fmt.Sprintf("%d %s errors", count, errorType.String()))
		}
	}
	
	return strings.Join(summary, ", ")
}

// AllMessages returns all error and warning messages
func (vr *ValidationResult) AllMessages() []string {
	var messages []string
	
	for _, err := range vr.Errors {
		messages = append(messages, "ERROR: "+err.Error())
	}
	
	for _, warning := range vr.Warning {
		messages = append(messages, "WARNING: "+warning)
	}
	
	return messages
}

// ErrorHandler provides utilities for handling and categorizing errors
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// CategorizeError categorizes a generic error into a ParseError
func (eh *ErrorHandler) CategorizeError(err error, context string) *ParseError {
	if err == nil {
		return nil
	}
	
	// Check if it's already a ParseError
	if parseErr, ok := err.(*ParseError); ok {
		if parseErr.Context == "" && context != "" {
			parseErr.Context = context
		}
		return parseErr
	}
	
	errMsg := err.Error()
	errMsgLower := strings.ToLower(errMsg)
	
	// Categorize based on error message patterns
	switch {
	case strings.Contains(errMsgLower, "no such file") || strings.Contains(errMsgLower, "not found"):
		return &ParseError{
			Type:     ErrorTypeFileNotFound,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	case strings.Contains(errMsgLower, "permission") || strings.Contains(errMsgLower, "access denied"):
		return &ParseError{
			Type:     ErrorTypeFilePermission,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	case strings.Contains(errMsgLower, "invalid json") || strings.Contains(errMsgLower, "invalid xml"):
		return &ParseError{
			Type:     ErrorTypeInvalidFormat,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	case strings.Contains(errMsgLower, "validation") || strings.Contains(errMsgLower, "required"):
		return &ParseError{
			Type:     ErrorTypeValidationFailed,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	case strings.Contains(errMsgLower, "corrupted") || strings.Contains(errMsgLower, "malformed"):
		return &ParseError{
			Type:     ErrorTypeCorruptedData,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	case strings.Contains(errMsgLower, "version") || strings.Contains(errMsgLower, "unsupported"):
		return &ParseError{
			Type:     ErrorTypeUnsupportedVersion,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	default:
		return &ParseError{
			Type:     ErrorTypeUnknown,
			Message:  errMsg,
			Context:  context,
			Original: err,
		}
	}
}

// ValidateFileAccess checks if a file can be accessed for reading
func (eh *ErrorHandler) ValidateFileAccess(filePath string) *ParseError {
	// This would typically check file permissions, existence, etc.
	// For now, we'll implement basic checks
	if filePath == "" {
		return NewParseError(ErrorTypeInvalidField, "file path cannot be empty")
	}
	
	// Additional file access validation could be added here
	return nil
}

// WrapError wraps an error with additional context
func (eh *ErrorHandler) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	
	if parseErr := eh.CategorizeError(err, context); parseErr != nil {
		return parseErr
	}
	
	return fmt.Errorf("%s: %w", context, err)
}

// RecoverFromPanic recovers from a panic and converts it to a ParseError
func (eh *ErrorHandler) RecoverFromPanic() *ParseError {
	if r := recover(); r != nil {
		return &ParseError{
			Type:    ErrorTypeCorruptedData,
			Message: fmt.Sprintf("parser panic: %v", r),
		}
	}
	return nil
}

// CreateValidationResult creates a new validation result
func (eh *ErrorHandler) CreateValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:   true,
		Errors:  []*ParseError{},
		Warning: []string{},
	}
}

// ValidateRequired validates that required fields are present
func (eh *ErrorHandler) ValidateRequired(fieldName, value string) *ParseError {
	if value == "" {
		return NewParseErrorWithField(ErrorTypeMissingField, 
			fmt.Sprintf("required field '%s' is missing or empty", fieldName), fieldName)
	}
	return nil
}

// ValidateUUID validates UUID format
func (eh *ErrorHandler) ValidateUUID(fieldName, uuid string) *ParseError {
	if uuid == "" {
		return nil // Allow empty UUIDs
	}
	
	// Basic UUID format validation (36 characters with 4 hyphens)
	if len(uuid) != 36 || strings.Count(uuid, "-") != 4 {
		return NewParseErrorWithField(ErrorTypeInvalidField,
			fmt.Sprintf("invalid UUID format: %s", uuid), fieldName)
	}
	
	return nil
}

// ValidateEnum validates that a value is in a set of allowed values
func (eh *ErrorHandler) ValidateEnum(fieldName, value string, allowedValues []string) *ParseError {
	if value == "" {
		return nil // Allow empty values
	}
	
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	
	return NewParseErrorWithField(ErrorTypeInvalidField,
		fmt.Sprintf("invalid value '%s' for field '%s', allowed values: %v", 
			value, fieldName, allowedValues), fieldName)
}