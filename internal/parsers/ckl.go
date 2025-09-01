package parsers

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthowerules/stigx/internal/models"
)

// CKLParser handles parsing of CKL (XML) format files
type CKLParser struct {
	securityValidator *models.SecurityValidator
}

// NewCKLParser creates a new CKL parser instance
func NewCKLParser() *CKLParser {
	return &CKLParser{
		securityValidator: models.NewSecurityValidator(),
	}
}

// ParseFile parses a CKL file from the given file path
func (p *CKLParser) ParseFile(filePath string) (*models.CKL, error) {
	if !p.IsCKLFile(filePath) {
		return nil, models.ValidationError{
			Field:   "file_extension",
			Message: "file is not a valid CKL file (expected .ckl extension)",
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, models.FileError{
			Operation: "open",
			FilePath:  filePath,
			Cause:     err,
		}
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses CKL data from an io.Reader
func (p *CKLParser) Parse(reader io.Reader) (*models.CKL, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, models.FileError{
			Operation: "read",
			FilePath:  "stream",
			Cause:     err,
		}
	}

	// Validate content for security issues
	if err := p.securityValidator.ValidateFileContent(data); err != nil {
		return nil, err
	}

	// Additional XML-specific security validation
	if err := p.securityValidator.ValidateXMLContent(data); err != nil {
		return nil, err
	}

	var ckl models.CKL
	if err := xml.Unmarshal(data, &ckl); err != nil {
		return nil, models.ParseError{
			File:    "stream",
			Message: fmt.Sprintf("failed to parse CKL XML: %v", err),
		}
	}

	// Perform basic validation
	if err := p.ValidateBasic(&ckl); err != nil {
		return nil, models.ValidationError{
			Field:   "ckl_structure",
			Message: fmt.Sprintf("CKL validation failed: %v", err),
		}
	}

	return &ckl, nil
}

// ValidateBasic performs basic structural validation of CKL data
func (p *CKLParser) ValidateBasic(ckl *models.CKL) error {
	// Validate asset information
	if ckl.Asset.AssetType == "" {
		return models.ValidationError{
			Field:   "ASSET_TYPE",
			Message: "missing required field",
		}
	}

	// Validate STIGs
	if len(ckl.STIGs.ISTIGs) == 0 {
		return models.ValidationError{
			Field:   "STIGs",
			Message: "no STIGs found in checklist",
		}
	}

	for i, istig := range ckl.STIGs.ISTIGs {
		if err := p.validateISTIG(&istig, i); err != nil {
			return err
		}
	}

	return nil
}

// validateISTIG validates individual STIG data
func (p *CKLParser) validateISTIG(istig *models.ISTIG, index int) error {
	// Validate STIG info
	stigMetadata := p.ExtractSTIGMetadata(&istig.STIGInfo)
	if stigMetadata.STIGID == "" {
		return fmt.Errorf("STIG[%d]: missing STIG ID", index)
	}

	if stigMetadata.Title == "" {
		return fmt.Errorf("STIG[%d]: missing STIG title", index)
	}

	// Validate vulnerabilities
	for j, vuln := range istig.Vulns {
		if err := p.validateVuln(&vuln, index, j); err != nil {
			return err
		}
	}

	return nil
}

// validateVuln validates individual vulnerability data
func (p *CKLParser) validateVuln(vuln *models.Vuln, stigIndex, vulnIndex int) error {
	vulnMetadata := p.ExtractVulnMetadata(vuln)

	if vulnMetadata.RuleID == "" {
		return fmt.Errorf("STIG[%d].Vuln[%d]: missing Rule ID", stigIndex, vulnIndex)
	}

	if vulnMetadata.VulnNum == "" {
		return fmt.Errorf("STIG[%d].Vuln[%d]: missing Vuln Num", stigIndex, vulnIndex)
	}

	// Validate status if present
	if vuln.Status != "" {
		validStatuses := map[string]bool{
			"NotAFinding": true, "Open": true, "Not_Reviewed": true, "Not_Applicable": true,
		}
		if !validStatuses[vuln.Status] {
			return fmt.Errorf("STIG[%d].Vuln[%d]: invalid status: %s", stigIndex, vulnIndex, vuln.Status)
		}
	}

	return nil
}

// ExtractSTIGMetadata extracts structured metadata from STIG info
func (p *CKLParser) ExtractSTIGMetadata(stigInfo *models.STIGInfo) models.STIGMetadata {
	metadata := models.STIGMetadata{}

	for _, data := range stigInfo.SIData {
		switch strings.ToLower(data.SIDName) {
		case "version":
			metadata.Version = data.SIDData
		case "classification":
			metadata.Classification = data.SIDData
		case "customname":
			metadata.CustomName = data.SIDData
		case "stigid":
			metadata.STIGID = data.SIDData
		case "description":
			metadata.Description = data.SIDData
		case "filename":
			metadata.Filename = data.SIDData
		case "releaseinfo":
			metadata.ReleaseInfo = data.SIDData
		case "title":
			metadata.Title = data.SIDData
		case "uuid":
			metadata.UUID = data.SIDData
		case "notice":
			metadata.Notice = data.SIDData
		case "source":
			metadata.Source = data.SIDData
		}
	}

	return metadata
}

// ExtractVulnMetadata extracts structured metadata from vulnerability data
func (p *CKLParser) ExtractVulnMetadata(vuln *models.Vuln) models.VulnMetadata {
	metadata := models.VulnMetadata{}

	for _, data := range vuln.SVuln {
		switch data.VulnAttribute {
		case "Vuln_Num":
			metadata.VulnNum = data.AttributeData
		case "Severity":
			metadata.Severity = data.AttributeData
		case "Group_Title":
			metadata.GroupTitle = data.AttributeData
		case "Rule_ID":
			metadata.RuleID = data.AttributeData
		case "Rule_Ver":
			metadata.RuleVer = data.AttributeData
		case "Rule_Title":
			metadata.RuleTitle = data.AttributeData
		case "Vuln_Discuss":
			metadata.VulnDiscuss = data.AttributeData
		case "IA_Controls":
			metadata.IAControls = data.AttributeData
		case "Check_Content":
			metadata.CheckContent = data.AttributeData
		case "Fix_Text":
			metadata.FixText = data.AttributeData
		case "False_Positives":
			metadata.FalsePositives = data.AttributeData
		case "False_Negatives":
			metadata.FalseNegatives = data.AttributeData
		case "Documentable":
			metadata.Documentable = data.AttributeData
		case "Mitigations":
			metadata.Mitigations = data.AttributeData
		case "Potential_Impact":
			metadata.PotentialImpacts = data.AttributeData
		case "Third_Party_Tools":
			metadata.ThirdPartyTools = data.AttributeData
		case "Mitigation_Control":
			metadata.MitigationControl = data.AttributeData
		case "Responsibility":
			metadata.Responsibility = data.AttributeData
		case "Security_Override_Guidance":
			metadata.SecurityOverrideGuidance = data.AttributeData
		case "Check_Content_Ref":
			metadata.CheckContentRef = data.AttributeData
		case "Weight":
			metadata.Weight = data.AttributeData
		case "Class":
			metadata.Class = data.AttributeData
		case "STIGRef":
			metadata.STIGRef = data.AttributeData
		case "TargetKey":
			metadata.TargetKey = data.AttributeData
		case "STIG_UUID":
			metadata.STIGUuid = data.AttributeData
		case "LEGACY_ID":
			if metadata.LegacyID == nil {
				metadata.LegacyID = []string{}
			}
			metadata.LegacyID = append(metadata.LegacyID, data.AttributeData)
		case "CCI_REF":
			if metadata.CCI == nil {
				metadata.CCI = []string{}
			}
			metadata.CCI = append(metadata.CCI, data.AttributeData)
		}
	}

	return metadata
}

// IsCKLFile checks if the given file path appears to be a CKL file
func (p *CKLParser) IsCKLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".ckl" || ext == ".xml"
}

// WriteFile writes CKL data to a file
func (p *CKLParser) WriteFile(ckl *models.CKL, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	return p.Write(ckl, file)
}

// Write writes CKL data to an io.Writer with XML formatting
func (p *CKLParser) Write(ckl *models.CKL, writer io.Writer) error {
	// Write XML header
	if _, err := writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")); err != nil {
		return fmt.Errorf("failed to write XML header: %w", err)
	}

	// Write STIG Viewer comment
	if _, err := writer.Write([]byte(`<!--STIG Viewer X-->` + "\n")); err != nil {
		return fmt.Errorf("failed to write comment: %w", err)
	}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "\t") // Pretty print with tabs

	if err := encoder.Encode(ckl); err != nil {
		return fmt.Errorf("failed to encode CKL XML: %w", err)
	}

	return nil
}

// GetStatistics returns statistics about the CKL data
func (p *CKLParser) GetStatistics(ckl *models.CKL) map[string]interface{} {
	stats := make(map[string]interface{})

	stats["asset_type"] = ckl.Asset.AssetType
	stats["host_name"] = ckl.Asset.HostName
	stats["host_ip"] = ckl.Asset.HostIP
	stats["stig_count"] = len(ckl.STIGs.ISTIGs)

	totalVulns := 0
	statusCounts := make(map[string]int)
	severityCounts := make(map[string]int)

	for _, istig := range ckl.STIGs.ISTIGs {
		totalVulns += len(istig.Vulns)

		for _, vuln := range istig.Vulns {
			if vuln.Status != "" {
				statusCounts[vuln.Status]++
			}

			vulnMeta := p.ExtractVulnMetadata(&vuln)
			if vulnMeta.Severity != "" {
				severityCounts[strings.ToLower(vulnMeta.Severity)]++
			}
		}
	}

	stats["total_vulnerabilities"] = totalVulns
	stats["status_counts"] = statusCounts
	stats["severity_counts"] = severityCounts

	return stats
}

// ConvertStatusToCKLb converts CKL status format to CKLb status format
func (p *CKLParser) ConvertStatusToCKLb(cklStatus string) string {
	switch cklStatus {
	case "NotAFinding":
		return "not_a_finding"
	case "Open":
		return "open"
	case "Not_Reviewed":
		return "not_reviewed"
	case "Not_Applicable":
		return "not_applicable"
	default:
		return strings.ToLower(cklStatus)
	}
}

// ConvertStatusFromCKLb converts CKLb status format to CKL status format
func (p *CKLParser) ConvertStatusFromCKLb(cklbStatus string) string {
	switch cklbStatus {
	case "not_a_finding":
		return "NotAFinding"
	case "open":
		return "Open"
	case "not_reviewed":
		return "Not_Reviewed"
	case "not_applicable":
		return "Not_Applicable"
	default:
		// Try to convert to CKL format (capitalize first letter)
		if len(cklbStatus) > 0 {
			return strings.ToUpper(cklbStatus[:1]) + cklbStatus[1:]
		}
		return cklbStatus
	}
}
