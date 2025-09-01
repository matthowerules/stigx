package models

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrorHandler provides utility functions for consistent error handling
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler instance
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// IsValidationError checks if an error is a validation error
func (h *ErrorHandler) IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}

// IsFileError checks if an error is a file operation error
func (h *ErrorHandler) IsFileError(err error) bool {
	var fileErr FileError
	return errors.As(err, &fileErr)
}

// IsDatabaseError checks if an error is a database operation error
func (h *ErrorHandler) IsDatabaseError(err error) bool {
	var dbErr DatabaseError
	return errors.As(err, &dbErr)
}

// IsNotFoundError checks if an error is a not found error
func (h *ErrorHandler) IsNotFoundError(err error) bool {
	var notFoundErr NotFoundError
	return errors.As(err, &notFoundErr)
}

// GetRootCause returns the root cause of an error by unwrapping
func (h *ErrorHandler) GetRootCause(err error) error {
	for err != nil {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			break
		}
		err = unwrapped
	}
	return err
}

// WrapError wraps an error with additional context
func (h *ErrorHandler) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// HandleError provides a centralized error handling strategy
func (h *ErrorHandler) HandleError(err error) (userMessage string, logMessage string, isRecoverable bool) {
	if err == nil {
		return "", "", true
	}

	// Check for specific error types
	switch e := err.(type) {
	case ValidationError:
		return fmt.Sprintf("Invalid input: %s", e.Message),
			fmt.Sprintf("Validation error for field '%s': %s", e.Field, e.Message),
			true
	case FileError:
		return "File operation failed. Please check the file and try again.",
			fmt.Sprintf("File %s failed for '%s': %v", e.Operation, e.FilePath, e.Cause),
			true
	case DatabaseError:
		return "Database operation failed. Please try again.",
			fmt.Sprintf("Database %s failed on table '%s': %v", e.Operation, e.Table, e.Cause),
			true
	case NotFoundError:
		return fmt.Sprintf("%s not found.", e.Resource),
			fmt.Sprintf("%s with ID '%v' not found", e.Resource, e.ID),
			true
	case ParseError:
		return "File format is invalid or corrupted.",
			fmt.Sprintf("Parse error in file '%s': %s", e.File, e.Message),
			true
	case ConversionError:
		return fmt.Sprintf("Conversion failed: %s", e.Message),
			fmt.Sprintf("Conversion from %s to %s failed: %s (cause: %v)", e.FromFormat, e.ToFormat, e.Message, e.Cause),
			true
	default:
		// Unknown error type
		return "An unexpected error occurred. Please try again.",
			fmt.Sprintf("Unexpected error: %v", err),
			false
	}
}

// SecurityValidator provides security validation functions
type SecurityValidator struct {
	allowedExtensions []string
	maxFileSize       int64
	blockedPaths      []string
}

// NewSecurityValidator creates a new security validator with default settings
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		allowedExtensions: []string{".ckl", ".cklb", ".json", ".xml", ".zip"},
		maxFileSize:       100 * 1024 * 1024, // 100MB
		blockedPaths:      []string{"/etc", "/usr", "/bin", "/sbin", "/dev", "/proc", "/sys", "C:\\Windows", "C:\\System32"},
	}
}

// ValidateFilePath validates a file path for security issues
func (sv *SecurityValidator) ValidateFilePath(filePath string) error {
	if filePath == "" {
		return ValidationError{
			Field:   "file_path",
			Message: "file path cannot be empty",
		}
	}

	// Check for path traversal attempts
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		return ValidationError{
			Field:   "file_path",
			Message: "path traversal detected",
		}
	}

	// Check for absolute paths that might access system files
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return FileError{
			Operation: "validate path",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	// Check against blocked system paths
	for _, blocked := range sv.blockedPaths {
		if strings.HasPrefix(strings.ToLower(absPath), strings.ToLower(blocked)) {
			return ValidationError{
				Field:   "file_path",
				Message: "access to system path is not allowed",
			}
		}
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	allowed := false
	for _, allowedExt := range sv.allowedExtensions {
		if ext == allowedExt {
			allowed = true
			break
		}
	}

	if !allowed {
		return ValidationError{
			Field:   "file_extension",
			Message: fmt.Sprintf("file extension '%s' is not allowed", ext),
		}
	}

	return nil
}

// ValidateFileContent performs basic content validation for security
func (sv *SecurityValidator) ValidateFileContent(content []byte) error {
	if len(content) == 0 {
		return ValidationError{
			Field:   "file_content",
			Message: "file content cannot be empty",
		}
	}

	if int64(len(content)) > sv.maxFileSize {
		return ValidationError{
			Field:   "file_size",
			Message: fmt.Sprintf("file size %d bytes exceeds maximum allowed size %d bytes", len(content), sv.maxFileSize),
		}
	}

	// Check for potentially malicious content patterns
	contentStr := string(content)

	// Check for script tags in XML/HTML content
	if strings.Contains(contentStr, "<script") || strings.Contains(contentStr, "javascript:") {
		return ValidationError{
			Field:   "file_content",
			Message: "potentially malicious script content detected",
		}
	}

	// Check for entity declarations that might be used for XXE attacks
	if strings.Contains(contentStr, "<!ENTITY") {
		return ValidationError{
			Field:   "file_content",
			Message: "XML entity declarations detected - potential XXE attack",
		}
	}

	// Check for null bytes that might indicate binary content in text files
	if strings.Contains(contentStr, "\x00") {
		return ValidationError{
			Field:   "file_content",
			Message: "null bytes detected in file content",
		}
	}

	return nil
}

// SanitizeFileName sanitizes a filename to prevent injection attacks
func (sv *SecurityValidator) SanitizeFileName(filename string) string {
	// Remove any path components
	filename = filepath.Base(filename)

	// Remove potentially dangerous characters
	dangerousChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = dangerousChars.ReplaceAllString(filename, "_")

	// Limit length
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		if len(nameWithoutExt) > 250 {
			nameWithoutExt = nameWithoutExt[:250]
		}
		filename = nameWithoutExt + ext
	}

	// Ensure it's not empty after sanitization
	if filename == "" || filename == "." || filename == ".." {
		filename = "unnamed_file"
	}

	return filename
}

// ValidateXMLContent performs XML-specific security validation
func (sv *SecurityValidator) ValidateXMLContent(content []byte) error {
	contentStr := strings.ToLower(string(content))

	// Check for XML processing instructions that might be dangerous
	if strings.Contains(contentStr, "<?xml-stylesheet") {
		return ValidationError{
			Field:   "xml_content",
			Message: "XML stylesheet processing instruction detected",
		}
	}

	// Check for DOCTYPE declarations that might enable XXE
	if strings.Contains(contentStr, "<!doctype") {
		return ValidationError{
			Field:   "xml_content",
			Message: "DOCTYPE declaration detected - potential XXE vulnerability",
		}
	}

	// Check for external entity references
	if strings.Contains(contentStr, "://") || strings.Contains(contentStr, "file://") {
		return ValidationError{
			Field:   "xml_content",
			Message: "external resource references detected",
		}
	}

	return nil
}

// ValidationError represents validation-specific errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// ParseError represents parsing-specific errors
type ParseError struct {
	File    string
	Line    int
	Message string
}

func (e ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error in file '%s' at line %d: %s", e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("parse error in file '%s': %s", e.File, e.Message)
}

// FileError represents file operation errors
type FileError struct {
	Operation string
	FilePath  string
	Cause     error
}

func (e FileError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("file %s failed for '%s': %v", e.Operation, e.FilePath, e.Cause)
	}
	return fmt.Sprintf("file %s failed for '%s'", e.Operation, e.FilePath)
}

func (e FileError) Unwrap() error {
	return e.Cause
}

// DatabaseError represents database operation errors
type DatabaseError struct {
	Operation string
	Table     string
	Cause     error
}

func (e DatabaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("database %s failed on table '%s': %v", e.Operation, e.Table, e.Cause)
	}
	return fmt.Sprintf("database %s failed on table '%s'", e.Operation, e.Table)
}

func (e DatabaseError) Unwrap() error {
	return e.Cause
}

// ConversionError represents format conversion errors
type ConversionError struct {
	FromFormat string
	ToFormat   string
	Message    string
	Cause      error
}

func (e ConversionError) Error() string {
	msg := fmt.Sprintf("conversion from %s to %s failed: %s", e.FromFormat, e.ToFormat, e.Message)
	if e.Cause != nil {
		msg += fmt.Sprintf(" (cause: %v)", e.Cause)
	}
	return msg
}

func (e ConversionError) Unwrap() error {
	return e.Cause
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID '%v' not found", e.Resource, e.ID)
}

// ConfigurationError represents configuration-related errors
type ConfigurationError struct {
	Setting string
	Message string
}

func (e ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error for '%s': %s", e.Setting, e.Message)
}

// FindingStatus represents the possible states of a STIG finding
type FindingStatus string

const (
	StatusNotReviewed   FindingStatus = "not_reviewed"
	StatusOpen          FindingStatus = "open"
	StatusNotAFinding   FindingStatus = "not_a_finding"
	StatusNotApplicable FindingStatus = "not_applicable"
)

// Severity represents the severity levels for STIG findings
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// ChecklistFormat represents the format of a checklist file
type ChecklistFormat string

const (
	FormatCKL   ChecklistFormat = "ckl"
	FormatCKLb  ChecklistFormat = "cklb"
	FormatXCCDF ChecklistFormat = "xccdf"
)

// Checklist represents a common interface for all checklist formats
type Checklist interface {
	GetTitle() string
	GetID() string
	GetSTIGs() []STIGInterface
	GetTargetData() TargetDataInterface
	SetTitle(string)
	SetID(string)
}

// STIGInterface represents a common interface for STIG data across formats
type STIGInterface interface {
	GetName() string
	GetID() string
	GetVersion() string
	GetReleaseInfo() string
	GetRules() []RuleInterface
}

// RuleInterface represents a common interface for STIG rules across formats
type RuleInterface interface {
	GetRuleID() string
	GetGroupID() string
	GetTitle() string
	GetSeverity() Severity
	GetStatus() FindingStatus
	GetComments() string
	GetFindingDetails() string
	SetStatus(FindingStatus)
	SetComments(string)
	SetFindingDetails(string)
}

// TargetDataInterface represents common target/asset information
type TargetDataInterface interface {
	GetHostName() string
	GetIPAddress() string
	GetRole() string
	SetHostName(string)
	SetIPAddress(string)
	SetRole(string)
}
