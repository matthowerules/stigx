package models

import (
	"strings"
	"testing"
)

func TestSecurityValidator_ValidateFilePath(t *testing.T) {
	validator := NewSecurityValidator()

	// Test valid file path
	validPath := "/home/user/test.ckl"
	err := validator.ValidateFilePath(validPath)
	if err != nil {
		t.Errorf("Expected valid path to pass validation, got error: %v", err)
	}

	// Test path traversal attempt
	invalidPath := "../../../etc/passwd"
	err = validator.ValidateFilePath(invalidPath)
	if err == nil {
		t.Error("Expected path traversal to fail validation")
	}

	// Test system path access
	systemPath := "/etc/passwd"
	err = validator.ValidateFilePath(systemPath)
	if err == nil {
		t.Error("Expected system path access to fail validation")
	}

	// Test invalid file extension
	invalidExt := "/home/user/test.exe"
	err = validator.ValidateFilePath(invalidExt)
	if err == nil {
		t.Error("Expected invalid file extension to fail validation")
	}

	// Test empty path
	emptyPath := ""
	err = validator.ValidateFilePath(emptyPath)
	if err == nil {
		t.Error("Expected empty path to fail validation")
	}
}

func TestSecurityValidator_ValidateFileContent(t *testing.T) {
	validator := NewSecurityValidator()

	// Test valid content
	validContent := []byte(`<?xml version="1.0"?><root><data>test</data></root>`)
	err := validator.ValidateFileContent(validContent)
	if err != nil {
		t.Errorf("Expected valid content to pass validation, got error: %v", err)
	}

	// Test empty content
	emptyContent := []byte{}
	err = validator.ValidateFileContent(emptyContent)
	if err == nil {
		t.Error("Expected empty content to fail validation")
	}

	// Test oversized content
	oversizedContent := make([]byte, 101*1024*1024) // 101MB
	err = validator.ValidateFileContent(oversizedContent)
	if err == nil {
		t.Error("Expected oversized content to fail validation")
	}

	// Test content with script tags
	scriptContent := []byte(`<script>alert('xss')</script>`)
	err = validator.ValidateFileContent(scriptContent)
	if err == nil {
		t.Error("Expected script content to fail validation")
	}

	// Test content with null bytes
	nullContent := []byte("test\x00data")
	err = validator.ValidateFileContent(nullContent)
	if err == nil {
		t.Error("Expected null byte content to fail validation")
	}
}

func TestSecurityValidator_SanitizeFileName(t *testing.T) {
	validator := NewSecurityValidator()

	// Test normal filename
	normalName := "test.ckl"
	sanitized := validator.SanitizeFileName(normalName)
	if sanitized != normalName {
		t.Errorf("Expected normal filename to remain unchanged, got: %s", sanitized)
	}

	// Test filename with dangerous characters
	dangerousName := "test<script>.ckl"
	sanitized = validator.SanitizeFileName(dangerousName)
	if strings.Contains(sanitized, "<") || strings.Contains(sanitized, ">") {
		t.Error("Expected dangerous characters to be sanitized")
	}

	// Test very long filename
	longName := strings.Repeat("a", 300) + ".ckl"
	sanitized = validator.SanitizeFileName(longName)
	if len(sanitized) > 255 {
		t.Error("Expected long filename to be truncated")
	}

	// Test empty filename
	emptyName := ""
	sanitized = validator.SanitizeFileName(emptyName)
	if sanitized == "" {
		t.Error("Expected empty filename to be replaced with default")
	}

	// Test path traversal in filename
	traversalName := "../../../evil.ckl"
	sanitized = validator.SanitizeFileName(traversalName)
	if strings.Contains(sanitized, "..") {
		t.Error("Expected path traversal to be sanitized")
	}
}

func TestSecurityValidator_ValidateXMLContent(t *testing.T) {
	validator := NewSecurityValidator()

	// Test valid XML content
	validXML := []byte(`<?xml version="1.0"?><root><data>test</data></root>`)
	err := validator.ValidateXMLContent(validXML)
	if err != nil {
		t.Errorf("Expected valid XML to pass validation, got error: %v", err)
	}

	// Test XML with stylesheet processing instruction
	stylesheetXML := []byte(`<?xml-stylesheet type="text/xsl" href="evil.xsl"?>`)
	err = validator.ValidateXMLContent(stylesheetXML)
	if err == nil {
		t.Error("Expected XML with stylesheet to fail validation")
	}

	// Test XML with DOCTYPE declaration
	doctypeXML := []byte(`<!DOCTYPE root SYSTEM "evil.dtd">`)
	err = validator.ValidateXMLContent(doctypeXML)
	if err == nil {
		t.Error("Expected XML with DOCTYPE to fail validation")
	}

	// Test XML with external entity reference
	externalXML := []byte(`<root>&entity;</root>`)
	err = validator.ValidateXMLContent(externalXML)
	if err == nil {
		t.Error("Expected XML with external entity to fail validation")
	}

	// Test XML with external URL reference
	urlXML := []byte(`<root xmlns:xi="http://www.w3.org/2001/XInclude"><xi:include href="http://evil.com"/></root>`)
	err = validator.ValidateXMLContent(urlXML)
	if err == nil {
		t.Error("Expected XML with external URL to fail validation")
	}
}
