package parsers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthowerules/stigx/internal/models"
)

// FileType represents the detected file format
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeCKL              // XML format
	FileTypeCKLb             // JSON format
	FileTypeXCCDF            // XCCDF format
)

// String returns the string representation of FileType
func (ft FileType) String() string {
	switch ft {
	case FileTypeCKL:
		return "CKL"
	case FileTypeCKLb:
		return "CKLb"
	case FileTypeXCCDF:
		return "XCCDF"
	default:
		return "Unknown"
	}
}

// FormatExtension returns the typical file extension for the format
func (ft FileType) FormatExtension() string {
	switch ft {
	case FileTypeCKL:
		return ".ckl"
	case FileTypeCKLb:
		return ".cklb"
	case FileTypeXCCDF:
		return ".xml"
	default:
		return ""
	}
}

// FileDetector provides file type detection capabilities
type FileDetector struct {
	securityValidator *models.SecurityValidator
}

// NewFileDetector creates a new file detector instance
func NewFileDetector() *FileDetector {
	return &FileDetector{
		securityValidator: models.NewSecurityValidator(),
	}
}

// DetectFile detects the format of a file by examining its contents
func (d *FileDetector) DetectFile(filePath string) (FileType, error) {
	// First validate the file path for security
	if err := d.securityValidator.ValidateFilePath(filePath); err != nil {
		return FileTypeUnknown, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return FileTypeUnknown, models.FileError{
			Operation: "open",
			FilePath:  filePath,
			Cause:     err,
		}
	}
	defer file.Close()

	return d.DetectWithSecurity(file, filePath)
}

// Detect detects the format by examining the content from an io.Reader
func (d *FileDetector) Detect(reader io.Reader) (FileType, error) {
	return d.DetectWithSecurity(reader, "")
}

// DetectWithSecurity detects the format with security validation
func (d *FileDetector) DetectWithSecurity(reader io.Reader, filePath string) (FileType, error) {
	// Read first chunk to analyze - increased size for better JSON detection
	buffer := make([]byte, 32768) // 32KB for large JSON files
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return FileTypeUnknown, models.FileError{
			Operation: "read",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	content := buffer[:n]

	// Validate content for security issues
	if err := d.securityValidator.ValidateFileContent(content); err != nil {
		return FileTypeUnknown, err
	}

	// Additional XML-specific validation if it looks like XML
	contentStr := string(content)
	if strings.HasPrefix(contentStr, "<?xml") || strings.HasPrefix(contentStr, "<") {
		if err := d.securityValidator.ValidateXMLContent(content); err != nil {
			return FileTypeUnknown, err
		}
	}

	return d.detectFromContent(contentStr), nil
}

// detectFromContent analyzes file content to determine format
func (d *FileDetector) detectFromContent(content string) FileType {
	// Remove leading/trailing whitespace
	content = strings.TrimSpace(content)

	// Check for JSON format (CKLb)
	if d.isJSONFormat(content) {
		return FileTypeCKLb
	}

	// Check for XML formats
	if d.isXMLFormat(content) {
		if d.isCKLFormat(content) {
			return FileTypeCKL
		}
		if d.isXCCDFFormat(content) {
			return FileTypeXCCDF
		}
	}

	return FileTypeUnknown
}

// isJSONFormat checks if content appears to be JSON
func (d *FileDetector) isJSONFormat(content string) bool {
	// Basic JSON structure check
	if !strings.HasPrefix(content, "{") {
		return false
	}

	// For detection, we'll skip full JSON parsing since large files may be truncated
	// Just check for CKLb-specific field patterns
	return (strings.Contains(content, `"title"`) && strings.Contains(content, `"stigs"`)) ||
		strings.Contains(content, `"cklb_version"`) ||
		(strings.Contains(content, `"stigs"`) && strings.Contains(content, `"stig_name"`)) ||
		(strings.Contains(content, `"stigs"`) && strings.Contains(content, `"rules"`)) ||
		(strings.Contains(content, `"stigs"`) && strings.Contains(content, `"display_name"`))
}

// isXMLFormat checks if content appears to be XML
func (d *FileDetector) isXMLFormat(content string) bool {
	return strings.HasPrefix(content, "<?xml") || strings.HasPrefix(content, "<")
}

// isCKLFormat checks if XML content is CKL format
func (d *FileDetector) isCKLFormat(content string) bool {
	// Check for CKL-specific root element and structure
	return strings.Contains(content, "<CHECKLIST>") &&
		strings.Contains(content, "<ASSET>") &&
		strings.Contains(content, "<STIGS>")
}

// isXCCDFFormat checks if XML content is XCCDF format
func (d *FileDetector) isXCCDFFormat(content string) bool {
	// Check for XCCDF-specific namespace and elements
	return (strings.Contains(content, "xccdf") || strings.Contains(content, "XCCDF")) &&
		(strings.Contains(content, "<Benchmark") || strings.Contains(content, "<benchmark"))
}

// DetectByExtension provides a quick detection based on file extension
func (d *FileDetector) DetectByExtension(filePath string) FileType {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".ckl":
		return FileTypeCKL
	case ".cklb":
		return FileTypeCKLb
	case ".json":
		// Could be CKLb, need content analysis
		return FileTypeUnknown
	case ".xml":
		// Could be CKL or XCCDF, need content analysis
		return FileTypeUnknown
	default:
		return FileTypeUnknown
	}
}

// ValidateFormat validates if a file matches the expected format
func (d *FileDetector) ValidateFormat(filePath string, expectedFormat FileType) error {
	detectedFormat, err := d.DetectFile(filePath)
	if err != nil {
		return models.FileError{
			Operation: "validate format",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	if detectedFormat != expectedFormat {
		return models.ValidationError{
			Field: "file_format",
			Message: fmt.Sprintf("expected %s, detected %s",
				expectedFormat.String(), detectedFormat.String()),
		}
	}

	return nil
}

// GetParser returns the appropriate parser for the detected file type
func (d *FileDetector) GetParser(fileType FileType) (interface{}, error) {
	switch fileType {
	case FileTypeCKL:
		return NewCKLParser(), nil
	case FileTypeCKLb:
		return NewCKLbParser(), nil
	default:
		return nil, models.ValidationError{
			Field:   "file_type",
			Message: fmt.Sprintf("no parser available for file type: %s", fileType.String()),
		}
	}
}

// AutoDetectAndParse automatically detects file format and returns parsed data
func (d *FileDetector) AutoDetectAndParse(filePath string) (interface{}, FileType, error) {
	fileType, err := d.DetectFile(filePath)
	if err != nil {
		return nil, FileTypeUnknown, models.FileError{
			Operation: "auto-detect and parse",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	switch fileType {
	case FileTypeCKL:
		parser := NewCKLParser()
		data, err := parser.ParseFile(filePath)
		return data, fileType, err

	case FileTypeCKLb:
		parser := NewCKLbParser()
		data, err := parser.ParseFile(filePath)
		return data, fileType, err

	default:
		return nil, fileType, models.ValidationError{
			Field:   "file_type",
			Message: fmt.Sprintf("unsupported file type: %s", fileType.String()),
		}
	}
}

// GetFileInfo returns detailed information about a file
func (d *FileDetector) GetFileInfo(filePath string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// File system info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, models.FileError{
			Operation: "stat",
			FilePath:  filePath,
			Cause:     err,
		}
	}

	info["file_name"] = filepath.Base(filePath)
	info["file_size"] = fileInfo.Size()
	info["modified_time"] = fileInfo.ModTime()
	info["extension"] = filepath.Ext(filePath)

	// Format detection
	detectedType, err := d.DetectFile(filePath)
	if err != nil {
		info["format_error"] = err.Error()
	} else {
		info["detected_format"] = detectedType.String()

		// Try to get format-specific statistics
		data, _, err := d.AutoDetectAndParse(filePath)
		if err != nil {
			info["parse_error"] = err.Error()
		} else {
			switch detectedType {
			case FileTypeCKL:
				if ckl, ok := data.(*models.CKL); ok {
					parser := NewCKLParser()
					info["statistics"] = parser.GetStatistics(ckl)
				}
			case FileTypeCKLb:
				if cklb, ok := data.(*models.CKLb); ok {
					parser := NewCKLbParser()
					info["statistics"] = parser.GetStatistics(cklb)
				}
			}
		}
	}

	return info, nil
}

// SuggestOutputFormat suggests an appropriate output format based on input
func (d *FileDetector) SuggestOutputFormat(inputPath, outputPath string) (FileType, error) {
	// If output path has a specific extension, use that
	outputExt := strings.ToLower(filepath.Ext(outputPath))
	switch outputExt {
	case ".ckl":
		return FileTypeCKL, nil
	case ".cklb", ".json":
		return FileTypeCKLb, nil
	}

	// If no specific extension, suggest opposite of input format
	inputType, err := d.DetectFile(inputPath)
	if err != nil {
		return FileTypeUnknown, models.FileError{
			Operation: "suggest output format",
			FilePath:  inputPath,
			Cause:     err,
		}
	}

	switch inputType {
	case FileTypeCKL:
		return FileTypeCKLb, nil
	case FileTypeCKLb:
		return FileTypeCKL, nil
	default:
		return FileTypeUnknown, models.ValidationError{
			Field:   "input_format",
			Message: fmt.Sprintf("cannot suggest output format for input type: %s", inputType.String()),
		}
	}
}
