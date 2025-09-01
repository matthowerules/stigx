package parsers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthowerules/stigx/internal/models"
)

// CKLbParser handles parsing of CKLb (JSON) format files
type CKLbParser struct{}

// NewCKLbParser creates a new CKLb parser instance
func NewCKLbParser() *CKLbParser {
	return &CKLbParser{}
}

// ParseFile parses a CKLb file from the given file path
func (p *CKLbParser) ParseFile(filePath string) (*models.CKLb, error) {
	if !p.IsCKLbFile(filePath) {
		return nil, fmt.Errorf("file %s is not a valid CKLb file (expected .cklb extension)", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses CKLb data from an io.Reader
func (p *CKLbParser) Parse(reader io.Reader) (*models.CKLb, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read CKLb data: %w", err)
	}

	var cklb models.CKLb
	if err := json.Unmarshal(data, &cklb); err != nil {
		return nil, fmt.Errorf("failed to parse CKLb JSON: %w", err)
	}

	// Perform basic validation
	if err := p.ValidateBasic(&cklb); err != nil {
		return nil, fmt.Errorf("CKLb validation failed: %w", err)
	}

	return &cklb, nil
}

// ValidateBasic performs basic structural validation of CKLb data
func (p *CKLbParser) ValidateBasic(cklb *models.CKLb) error {
	// Some CKLb files may not have title or ID at the top level, generate defaults
	if cklb.Title == "" && len(cklb.STIGs) > 0 {
		cklb.Title = cklb.STIGs[0].STIGName + " Checklist"
	}
	if cklb.Title == "" {
		cklb.Title = "Untitled Checklist"
	}
	
	if cklb.ID == "" {
		// Generate a placeholder ID if missing
		cklb.ID = "00000000-0000-0000-0000-000000000000"
	}

	// Validate UUID format (basic check) only if not placeholder
	if cklb.ID != "00000000-0000-0000-0000-000000000000" && 
	   (len(cklb.ID) != 36 || strings.Count(cklb.ID, "-") != 4) {
		return fmt.Errorf("invalid UUID format for id field: %s", cklb.ID)
	}

	// Validate STIGs if present
	for i, stig := range cklb.STIGs {
		if err := p.validateSTIG(&stig, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSTIG validates individual STIG data
func (p *CKLbParser) validateSTIG(stig *models.STIG, index int) error {
	if stig.STIGName == "" {
		return fmt.Errorf("STIG[%d]: missing required field: stig_name", index)
	}
	
	if stig.DisplayName == "" {
		return fmt.Errorf("STIG[%d]: missing required field: display_name", index)
	}
	
	if stig.STIGID == "" {
		return fmt.Errorf("STIG[%d]: missing required field: stig_id", index)
	}
	
	if stig.ReleaseInfo == "" {
		return fmt.Errorf("STIG[%d]: missing required field: release_info", index)
	}
	
	if stig.UUID == "" {
		return fmt.Errorf("STIG[%d]: missing required field: uuid", index)
	}

	// Validate UUID format
	if len(stig.UUID) != 36 || strings.Count(stig.UUID, "-") != 4 {
		return fmt.Errorf("STIG[%d]: invalid UUID format: %s", index, stig.UUID)
	}

	if stig.Size <= 0 {
		return fmt.Errorf("STIG[%d]: size must be greater than 0, got: %d", index, stig.Size)
	}

	// Validate rules
	for j, rule := range stig.Rules {
		if err := p.validateSTIGRule(&rule, index, j); err != nil {
			return err
		}
	}

	return nil
}

// validateSTIGRule validates individual STIG rule data
func (p *CKLbParser) validateSTIGRule(rule *models.STIGRule, stigIndex, ruleIndex int) error {
	if rule.UUID == "" {
		return fmt.Errorf("STIG[%d].Rule[%d]: missing required field: uuid", stigIndex, ruleIndex)
	}
	
	if rule.STIGUUID == "" {
		return fmt.Errorf("STIG[%d].Rule[%d]: missing required field: stig_uuid", stigIndex, ruleIndex)
	}
	
	if rule.GroupID == "" {
		return fmt.Errorf("STIG[%d].Rule[%d]: missing required field: group_id", stigIndex, ruleIndex)
	}
	
	if rule.RuleID == "" {
		return fmt.Errorf("STIG[%d].Rule[%d]: missing required field: rule_id", stigIndex, ruleIndex)
	}

	// Validate severity
	validSeverities := map[string]bool{
		"high": true, "medium": true, "low": true,
	}
	if rule.Severity != "" && !validSeverities[strings.ToLower(rule.Severity)] {
		return fmt.Errorf("STIG[%d].Rule[%d]: invalid severity: %s (must be high, medium, or low)", 
			stigIndex, ruleIndex, rule.Severity)
	}

	// Validate status
	validStatuses := map[string]bool{
		"not_reviewed": true, "open": true, "not_a_finding": true, "not_applicable": true,
	}
	if rule.Status != "" && !validStatuses[rule.Status] {
		return fmt.Errorf("STIG[%d].Rule[%d]: invalid status: %s", stigIndex, ruleIndex, rule.Status)
	}

	return nil
}

// IsCKLbFile checks if the given file path appears to be a CKLb file
func (p *CKLbParser) IsCKLbFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".cklb" || ext == ".json"
}

// WriteFile writes CKLb data to a file
func (p *CKLbParser) WriteFile(cklb *models.CKLb, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	return p.Write(cklb, file)
}

// Write writes CKLb data to an io.Writer with pretty formatting
func (p *CKLbParser) Write(cklb *models.CKLb, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Pretty print with 2 spaces
	
	if err := encoder.Encode(cklb); err != nil {
		return fmt.Errorf("failed to encode CKLb JSON: %w", err)
	}

	return nil
}

// GetStatistics returns statistics about the CKLb data
func (p *CKLbParser) GetStatistics(cklb *models.CKLb) map[string]interface{} {
	stats := make(map[string]interface{})
	
	stats["title"] = cklb.Title
	stats["id"] = cklb.ID
	stats["stig_count"] = len(cklb.STIGs)
	
	totalRules := 0
	statusCounts := make(map[string]int)
	severityCounts := make(map[string]int)
	
	for _, stig := range cklb.STIGs {
		totalRules += len(stig.Rules)
		
		for _, rule := range stig.Rules {
			if rule.Status != "" {
				statusCounts[rule.Status]++
			}
			if rule.Severity != "" {
				severityCounts[strings.ToLower(rule.Severity)]++
			}
		}
	}
	
	stats["total_rules"] = totalRules
	stats["status_counts"] = statusCounts
	stats["severity_counts"] = severityCounts
	
	if cklb.TargetData != nil {
		stats["target_hostname"] = cklb.TargetData.HostName
		stats["target_ip"] = cklb.TargetData.IPAddress
	}
	
	return stats
}
