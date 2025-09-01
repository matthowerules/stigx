package converters

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matthowerules/stigx/internal/models"
	"github.com/matthowerules/stigx/internal/parsers"
)

// Converter handles conversion between CKL and CKLb formats
type Converter struct {
	cklParser  *parsers.CKLParser
	cklbParser *parsers.CKLbParser
}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	return &Converter{
		cklParser:  parsers.NewCKLParser(),
		cklbParser: parsers.NewCKLbParser(),
	}
}

// ConvertFile converts a file from one format to another
func (c *Converter) ConvertFile(inputPath, outputPath string) error {
	detector := parsers.NewFileDetector()

	// Detect input format
	inputType, err := detector.DetectFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect input format: %w", err)
	}

	// Suggest output format
	outputType, err := detector.SuggestOutputFormat(inputPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to determine output format: %w", err)
	}

	return c.ConvertFileWithFormats(inputPath, outputPath, inputType, outputType)
}

// ConvertFileWithFormats converts a file between specified formats
func (c *Converter) ConvertFileWithFormats(inputPath, outputPath string, inputType, outputType parsers.FileType) error {
	switch {
	case inputType == parsers.FileTypeCKL && outputType == parsers.FileTypeCKLb:
		return c.ConvertCKLToCKLb(inputPath, outputPath)
	case inputType == parsers.FileTypeCKLb && outputType == parsers.FileTypeCKL:
		return c.ConvertCKLbToCKL(inputPath, outputPath)
	default:
		return fmt.Errorf("unsupported conversion: %s to %s", inputType.String(), outputType.String())
	}
}

// ConvertCKLToCKLb converts a CKL file to CKLb format
func (c *Converter) ConvertCKLToCKLb(inputPath, outputPath string) error {
	// Parse CKL file
	ckl, err := c.cklParser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse CKL file: %w", err)
	}

	// Convert to CKLb
	cklb, err := c.CKLToCKLb(ckl)
	if err != nil {
		return fmt.Errorf("failed to convert CKL to CKLb: %w", err)
	}

	// Write CKLb file
	if err := c.cklbParser.WriteFile(cklb, outputPath); err != nil {
		return fmt.Errorf("failed to write CKLb file: %w", err)
	}

	return nil
}

// ConvertCKLbToCKL converts a CKLb file to CKL format
func (c *Converter) ConvertCKLbToCKL(inputPath, outputPath string) error {
	// Parse CKLb file
	cklb, err := c.cklbParser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to parse CKLb file: %w", err)
	}

	// Convert to CKL
	ckl, err := c.CKLbToCKL(cklb)
	if err != nil {
		return fmt.Errorf("failed to convert CKLb to CKL: %w", err)
	}

	// Write CKL file
	if err := c.cklParser.WriteFile(ckl, outputPath); err != nil {
		return fmt.Errorf("failed to write CKL file: %w", err)
	}

	return nil
}

// CKLToCKLb converts CKL data structure to CKLb
func (c *Converter) CKLToCKLb(ckl *models.CKL) (*models.CKLb, error) {
	cklb := &models.CKLb{
		ID:          generateUUIDIfEmpty(""),
		CklbVersion: "1.0",
	}

	// Convert target data from asset
	if ckl.Asset.AssetType != "" {
		cklb.TargetData = &models.TargetData{
			TargetType:     ckl.Asset.AssetType,
			HostName:       ckl.Asset.HostName,
			IPAddress:      ckl.Asset.HostIP,
			MACAddress:     ckl.Asset.HostMAC,
			FQDN:           ckl.Asset.HostFQDN,
			Comments:       ckl.Asset.TargetComment,
			Role:           ckl.Asset.Role,
			TechnologyArea: ckl.Asset.TechArea,
			WebDBSite:      ckl.Asset.WebDBSite,
			WebDBInstance:  ckl.Asset.WebDBInstance,
		}

		// Convert web database flag
		if ckl.Asset.WebOrDatabase == "true" {
			webDB := true
			cklb.TargetData.IsWebDatabase = &webDB
		}
	}

	// Convert STIGs
	for _, istig := range ckl.STIGs.ISTIGs {
		stigMeta := c.cklParser.ExtractSTIGMetadata(&istig.STIGInfo)

		stig := models.STIG{
			STIGName:    stigMeta.Title,
			DisplayName: c.createDisplayName(stigMeta.Title),
			STIGID:      stigMeta.STIGID,
			ReleaseInfo: stigMeta.ReleaseInfo,
			UUID:        generateUUIDIfEmpty(stigMeta.UUID),
			Size:        len(istig.Vulns),
			Version:     stigMeta.Version,
		}

		// Set title as fallback
		if stig.STIGName == "" {
			stig.STIGName = stigMeta.Description
		}
		if stig.STIGName == "" {
			stig.STIGName = "Unknown STIG"
		}

		// Set display name
		if stig.DisplayName == "" {
			stig.DisplayName = stig.STIGName
		}

		// Set reference identifier
		if ckl.Asset.TargetKey != "" {
			stig.ReferenceIdentifier = &ckl.Asset.TargetKey
		}

		// Convert vulnerabilities to rules
		for _, vuln := range istig.Vulns {
			rule, err := c.vulnToSTIGRule(&vuln, stig.UUID)
			if err != nil {
				return nil, fmt.Errorf("failed to convert vulnerability: %w", err)
			}
			stig.Rules = append(stig.Rules, *rule)
		}

		cklb.STIGs = append(cklb.STIGs, stig)
	}

	// Set title based on first STIG or default
	if len(cklb.STIGs) > 0 {
		cklb.Title = cklb.STIGs[0].DisplayName + " Checklist"
	} else {
		cklb.Title = "Converted Checklist"
	}

	return cklb, nil
}

// CKLbToCKL converts CKLb data structure to CKL
func (c *Converter) CKLbToCKL(cklb *models.CKLb) (*models.CKL, error) {
	ckl := &models.CKL{}

	// Convert target data to asset
	if cklb.TargetData != nil {
		ckl.Asset = models.Asset{
			AssetType:     cklb.TargetData.TargetType,
			HostName:      cklb.TargetData.HostName,
			HostIP:        cklb.TargetData.IPAddress,
			HostMAC:       cklb.TargetData.MACAddress,
			HostFQDN:      cklb.TargetData.FQDN,
			TargetComment: cklb.TargetData.Comments,
			Role:          cklb.TargetData.Role,
			TechArea:      cklb.TargetData.TechnologyArea,
			WebDBSite:     cklb.TargetData.WebDBSite,
			WebDBInstance: cklb.TargetData.WebDBInstance,
		}

		// Convert web database flag
		if cklb.TargetData.IsWebDatabase != nil && *cklb.TargetData.IsWebDatabase {
			ckl.Asset.WebOrDatabase = "true"
		} else {
			ckl.Asset.WebOrDatabase = "false"
		}

		// Set target key from reference identifier
		if len(cklb.STIGs) > 0 && cklb.STIGs[0].ReferenceIdentifier != nil {
			ckl.Asset.TargetKey = *cklb.STIGs[0].ReferenceIdentifier
		}
	}

	// Set default asset type if empty
	if ckl.Asset.AssetType == "" {
		ckl.Asset.AssetType = "Computing"
	}

	// Convert STIGs
	for _, stig := range cklb.STIGs {
		istig := models.ISTIG{
			STIGInfo: c.createSTIGInfo(stig),
		}

		// Convert rules to vulnerabilities
		for _, rule := range stig.Rules {
			vuln := c.stigRuleToVuln(&rule)
			istig.Vulns = append(istig.Vulns, *vuln)
		}

		ckl.STIGs.ISTIGs = append(ckl.STIGs.ISTIGs, istig)
	}

	return ckl, nil
}

// vulnToSTIGRule converts a CKL vulnerability to a CKLb STIG rule
func (c *Converter) vulnToSTIGRule(vuln *models.Vuln, stigUUID string) (*models.STIGRule, error) {
	vulnMeta := c.cklParser.ExtractVulnMetadata(vuln)

	rule := &models.STIGRule{
		UUID:                     generateUUIDIfEmpty(""),
		STIGUUID:                 stigUUID,
		STIGUuid:                 stigUUID, // Note: different casing
		GroupID:                  vulnMeta.VulnNum,
		GroupIDSrc:               vulnMeta.VulnNum,
		RuleID:                   vulnMeta.RuleID,
		RuleIDSrc:                vulnMeta.RuleID,
		Weight:                   vulnMeta.Weight,
		Classification:           vulnMeta.Class,
		Severity:                 strings.ToLower(vulnMeta.Severity), // WHY DOES THIS NOT WORK???
		RuleVersion:              vulnMeta.RuleVer,
		GroupTitle:               vulnMeta.GroupTitle,
		RuleTitle:                vulnMeta.RuleTitle,
		FixText:                  vulnMeta.FixText,
		FalsePositives:           vulnMeta.FalsePositives,
		FalseNegatives:           vulnMeta.FalseNegatives,
		Discussion:               vulnMeta.VulnDiscuss,
		CheckContent:             vulnMeta.CheckContent,
		Documentable:             vulnMeta.Documentable,
		Mitigations:              vulnMeta.Mitigations,
		PotentialImpacts:         vulnMeta.PotentialImpacts,
		ThirdPartyTools:          vulnMeta.ThirdPartyTools,
		MitigationControl:        vulnMeta.MitigationControl,
		Responsibility:           vulnMeta.Responsibility,
		SecurityOverrideGuidance: vulnMeta.SecurityOverrideGuidance,
		IAControls:               vulnMeta.IAControls,
		LegacyIDs:                vulnMeta.LegacyID,
		CCIs:                     vulnMeta.CCI,
		Status:                   c.cklParser.ConvertStatusToCKLb(vuln.Status),
		Comments:                 vuln.Comments,
		FindingDetails:           vuln.FindingDetails,
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		Overrides:                make(map[string]interface{}),
	}

	// Set target key and STIG ref
	if vulnMeta.TargetKey != "" {
		rule.TargetKey = &vulnMeta.TargetKey
	}
	if vulnMeta.STIGRef != "" {
		rule.STIGRef = &vulnMeta.STIGRef
	}

	// Create check content ref if available
	if vulnMeta.CheckContentRef != "" {
		rule.CheckContentRef = &models.ContentRef{
			Href: vulnMeta.CheckContentRef,
			Name: "M", // Default name
		}
	}

	// Create group tree
	if rule.GroupID != "" {
		rule.GroupTree = []models.GroupTree{
			{
				ID:          rule.GroupID,
				Title:       rule.GroupTitle,
				Description: "<GroupDescription></GroupDescription>",
			},
		}
	}

	return rule, nil
}

// stigRuleToVuln converts a CKLb STIG rule to a CKL vulnerability
func (c *Converter) stigRuleToVuln(rule *models.STIGRule) *models.Vuln {
	vuln := &models.Vuln{
		Status:                c.cklParser.ConvertStatusFromCKLb(rule.Status),
		FindingDetails:        rule.FindingDetails,
		Comments:              rule.Comments,
		SeverityOverride:      "", // Not directly mapped
		SeverityJustification: "", // Not directly mapped
	}

	// Build STIG data attributes
	vulnAttrs := []models.SVuln{
		{VulnAttribute: "Vuln_Num", AttributeData: rule.GroupID},
		{VulnAttribute: "Severity", AttributeData: strings.ToLower(rule.Severity)},
		{VulnAttribute: "Group_Title", AttributeData: rule.GroupTitle},
		{VulnAttribute: "Rule_ID", AttributeData: rule.RuleID},
		{VulnAttribute: "Rule_Ver", AttributeData: rule.RuleVersion},
		{VulnAttribute: "Rule_Title", AttributeData: rule.RuleTitle},
		{VulnAttribute: "Vuln_Discuss", AttributeData: rule.Discussion},
		{VulnAttribute: "IA_Controls", AttributeData: rule.IAControls},
		{VulnAttribute: "Check_Content", AttributeData: rule.CheckContent},
		{VulnAttribute: "Fix_Text", AttributeData: rule.FixText},
		{VulnAttribute: "False_Positives", AttributeData: rule.FalsePositives},
		{VulnAttribute: "False_Negatives", AttributeData: rule.FalseNegatives},
		{VulnAttribute: "Documentable", AttributeData: rule.Documentable},
		{VulnAttribute: "Mitigations", AttributeData: rule.Mitigations},
		{VulnAttribute: "Potential_Impact", AttributeData: rule.PotentialImpacts},
		{VulnAttribute: "Third_Party_Tools", AttributeData: rule.ThirdPartyTools},
		{VulnAttribute: "Mitigation_Control", AttributeData: rule.MitigationControl},
		{VulnAttribute: "Responsibility", AttributeData: rule.Responsibility},
		{VulnAttribute: "Security_Override_Guidance", AttributeData: rule.SecurityOverrideGuidance},
		{VulnAttribute: "Weight", AttributeData: rule.Weight},
		{VulnAttribute: "Class", AttributeData: rule.Classification},
		{VulnAttribute: "STIG_UUID", AttributeData: rule.STIGUUID},
	}

	// Add target key and STIG ref if available
	if rule.TargetKey != nil {
		vulnAttrs = append(vulnAttrs, models.SVuln{
			VulnAttribute: "TargetKey",
			AttributeData: *rule.TargetKey,
		})
	}
	if rule.STIGRef != nil {
		vulnAttrs = append(vulnAttrs, models.SVuln{
			VulnAttribute: "STIGRef",
			AttributeData: *rule.STIGRef,
		})
	}

	// Add check content ref
	if rule.CheckContentRef != nil {
		vulnAttrs = append(vulnAttrs, models.SVuln{
			VulnAttribute: "Check_Content_Ref",
			AttributeData: rule.CheckContentRef.Href,
		})
	}

	// Add legacy IDs
	for _, legacyID := range rule.LegacyIDs {
		vulnAttrs = append(vulnAttrs, models.SVuln{
			VulnAttribute: "LEGACY_ID",
			AttributeData: legacyID,
		})
	}

	// Add CCIs
	for _, cci := range rule.CCIs {
		vulnAttrs = append(vulnAttrs, models.SVuln{
			VulnAttribute: "CCI_REF",
			AttributeData: cci,
		})
	}

	vuln.SVuln = vulnAttrs
	return vuln
}

// createSTIGInfo creates STIG info structure from STIG data
func (c *Converter) createSTIGInfo(stig models.STIG) models.STIGInfo {
	return models.STIGInfo{
		SIData: []models.SIData{
			{SIDName: "version", SIDData: stig.Version},
			{SIDName: "classification", SIDData: "UNCLASSIFIED"},
			{SIDName: "customname", SIDData: ""},
			{SIDName: "stigid", SIDData: stig.STIGID},
			{SIDName: "description", SIDData: stig.STIGName},
			{SIDName: "filename", SIDData: stig.STIGID + ".xml"},
			{SIDName: "releaseinfo", SIDData: stig.ReleaseInfo},
			{SIDName: "title", SIDData: stig.STIGName},
			{SIDName: "uuid", SIDData: stig.UUID},
			{SIDName: "notice", SIDData: ""},
			{SIDName: "source", SIDData: "STIG"},
		},
	}
}

// createDisplayName creates a shortened display name from full STIG name
func (c *Converter) createDisplayName(fullName string) string {
	// Common abbreviations for display names
	displayName := fullName
	displayName = strings.ReplaceAll(displayName, "Security Technical Implementation Guide", "STIG")
	displayName = strings.ReplaceAll(displayName, "Technical Implementation Guide", "STIG")

	// Truncate if too long
	if len(displayName) > 80 {
		displayName = displayName[:77] + "..."
	}

	return displayName
}

// generateUUIDIfEmpty generates a new UUID if the input is empty
func generateUUIDIfEmpty(input string) string {
	if input == "" {
		return uuid.New().String()
	}
	return input
}
