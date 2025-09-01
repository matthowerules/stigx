package parsers

import (
	"testing"
)

func TestFileType_String(t *testing.T) {
	tests := []struct {
		fileType FileType
		expected string
	}{
		{FileTypeCKL, "CKL"},
		{FileTypeCKLb, "CKLb"},
		{FileTypeXCCDF, "XCCDF"},
		{FileTypeUnknown, "Unknown"},
	}

	for _, test := range tests {
		result := test.fileType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestFileType_FormatExtension(t *testing.T) {
	tests := []struct {
		fileType FileType
		expected string
	}{
		{FileTypeCKL, ".ckl"},
		{FileTypeCKLb, ".cklb"},
		{FileTypeXCCDF, ".xml"},
		{FileTypeUnknown, ""},
	}

	for _, test := range tests {
		result := test.fileType.FormatExtension()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestFileDetector_DetectByExtension(t *testing.T) {
	detector := NewFileDetector()

	tests := []struct {
		filePath string
		expected FileType
	}{
		{"test.ckl", FileTypeCKL},
		{"test.cklb", FileTypeCKLb},
		{"test.json", FileTypeUnknown}, // JSON needs content analysis
		{"test.xml", FileTypeUnknown},  // XML needs content analysis
		{"test.txt", FileTypeUnknown},
		{"", FileTypeUnknown},
	}

	for _, test := range tests {
		result := detector.DetectByExtension(test.filePath)
		if result != test.expected {
			t.Errorf("For path '%s', expected %s, got %s", test.filePath, test.expected, result)
		}
	}
}

func TestFileDetector_detectFromContent(t *testing.T) {
	detector := NewFileDetector()

	// Test CKLb detection
	cklbContent := `{"cklb_version": "1.0", "title": "Test Checklist", "stigs": []}`
	result := detector.detectFromContent(cklbContent)
	if result != FileTypeCKLb {
		t.Errorf("Expected CKLb, got %s", result)
	}

	// Test CKL detection
	cklContent := `<?xml version="1.0"?><CHECKLIST><ASSET><ASSET_TYPE>Computing</ASSET_TYPE></ASSET><STIGS></STIGS></CHECKLIST>`
	result = detector.detectFromContent(cklContent)
	if result != FileTypeCKL {
		t.Errorf("Expected CKL, got %s", result)
	}

	// Test XCCDF detection
	xccdfContent := `<?xml version="1.0"?><Benchmark xmlns="http://checklists.nist.gov/xccdf/1.2" id="test"><title>Test Benchmark</title></Benchmark>`
	result = detector.detectFromContent(xccdfContent)
	if result != FileTypeXCCDF {
		t.Errorf("Expected XCCDF, got %s", result)
	}

	// Test unknown content
	unknownContent := "This is not a valid STIG file format"
	result = detector.detectFromContent(unknownContent)
	if result != FileTypeUnknown {
		t.Errorf("Expected Unknown, got %s", result)
	}
}

func TestFileDetector_SuggestOutputFormat(t *testing.T) {
	detector := NewFileDetector()

	// Test CKL to CKLb conversion
	result, err := detector.SuggestOutputFormat("input.ckl", "output.cklb")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != FileTypeCKLb {
		t.Errorf("Expected CKLb, got %s", result)
	}

	// Test CKLb to CKL conversion
	result, err = detector.SuggestOutputFormat("input.cklb", "output.ckl")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != FileTypeCKL {
		t.Errorf("Expected CKL, got %s", result)
	}

	// Test with extension in output path
	result, err = detector.SuggestOutputFormat("input.ckl", "output.cklb")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != FileTypeCKLb {
		t.Errorf("Expected CKLb, got %s", result)
	}

	// Test unknown input type
	result, err = detector.SuggestOutputFormat("input.unknown", "output.ckl")
	if err == nil {
		t.Error("Expected error for unknown input type")
	}
}

func TestFileDetector_isJSONFormat(t *testing.T) {
	detector := NewFileDetector()

	// Test valid JSON
	validJSON := `{"title": "Test", "stigs": []}`
	if !detector.isJSONFormat(validJSON) {
		t.Error("Expected valid JSON to be detected")
	}

	// Test invalid JSON
	invalidJSON := "This is not JSON"
	if detector.isJSONFormat(invalidJSON) {
		t.Error("Expected invalid JSON to not be detected")
	}

	// Test empty string
	emptyString := ""
	if detector.isJSONFormat(emptyString) {
		t.Error("Expected empty string to not be detected as JSON")
	}

	// Test whitespace only
	whitespaceOnly := "   \n\t   "
	if detector.isJSONFormat(whitespaceOnly) {
		t.Error("Expected whitespace-only string to not be detected as JSON")
	}
}

func TestFileDetector_isXMLFormat(t *testing.T) {
	detector := NewFileDetector()

	// Test XML declaration
	xmlDecl := "<?xml version=\"1.0\"?>"
	if !detector.isXMLFormat(xmlDecl) {
		t.Error("Expected XML declaration to be detected")
	}

	// Test XML element
	xmlElement := "<root>"
	if !detector.isXMLFormat(xmlElement) {
		t.Error("Expected XML element to be detected")
	}

	// Test non-XML content
	nonXML := "This is not XML"
	if detector.isXMLFormat(nonXML) {
		t.Error("Expected non-XML content to not be detected")
	}
}

func TestFileDetector_isCKLFormat(t *testing.T) {
	detector := NewFileDetector()

	// Test valid CKL structure
	validCKL := `<?xml version="1.0"?><CHECKLIST><ASSET><ASSET_TYPE>Computing</ASSET_TYPE></ASSET><STIGS><STIG></STIG></STIGS></CHECKLIST>`
	if !detector.isCKLFormat(validCKL) {
		t.Error("Expected valid CKL structure to be detected")
	}

	// Test invalid CKL structure
	invalidCKL := `<?xml version="1.0"?><root><data>test</data></root>`
	if detector.isCKLFormat(invalidCKL) {
		t.Error("Expected invalid CKL structure to not be detected")
	}
}

func TestFileDetector_isXCCDFFormat(t *testing.T) {
	detector := NewFileDetector()

	// Test valid XCCDF structure
	validXCCDF := `<?xml version="1.0"?><Benchmark xmlns="http://checklists.nist.gov/xccdf/1.2" id="test"><title>Test</title></Benchmark>`
	if !detector.isXCCDFFormat(validXCCDF) {
		t.Error("Expected valid XCCDF structure to be detected")
	}

	// Test invalid XCCDF structure
	invalidXCCDF := `<?xml version="1.0"?><root><data>test</data></root>`
	if detector.isXCCDFFormat(invalidXCCDF) {
		t.Error("Expected invalid XCCDF structure to not be detected")
	}
}
