package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matthowerules/stigx/internal/database"
	"github.com/matthowerules/stigx/internal/models"
)

// ChecklistService handles checklist-related operations
type ChecklistService struct {
	db          *database.DB
	stigService *STIGService
}

// NewChecklistService creates a new checklist service
func NewChecklistService(db *database.DB, stigService *STIGService) *ChecklistService {
	return &ChecklistService{
		db:          db,
		stigService: stigService,
	}
}

// extractVulnDiscussionText extracts the text content from VulnDiscussion XML tags
func extractVulnDiscussionText(description string) string {
	if description == "" {
		return ""
	}

	// Look for VulnDiscussion tags (case insensitive)
	vulnDiscussionRegex := regexp.MustCompile(`(?i)<VulnDiscussion[^>]*>(.*?)</VulnDiscussion>`)
	matches := vulnDiscussionRegex.FindStringSubmatch(description)

	if len(matches) > 1 {
		// Return the content inside the VulnDiscussion tags
		content := strings.TrimSpace(matches[1])
		// Remove any remaining XML tags within the content
		cleanRegex := regexp.MustCompile(`<[^>]+>`)
		cleaned := cleanRegex.ReplaceAllString(content, "")
		return strings.TrimSpace(cleaned)
	}

	// If no VulnDiscussion tags found, return the original description
	// but clean up any other XML tags that might be present
	cleanRegex := regexp.MustCompile(`<[^>]+>`)
	cleaned := cleanRegex.ReplaceAllString(description, "")
	return strings.TrimSpace(cleaned)
}

// Example test cases for extractVulnDiscussionText:
// Input: "<VulnDiscussion>This is a test discussion</VulnDiscussion>"
// Output: "This is a test discussion"
//
// Input: "<VulnDiscussion><p>This is <b>important</b> information</p></VulnDiscussion>"
// Output: "This is important information"
//
// Input: "Some plain text without tags"
// Output: "Some plain text without tags"

// CreateChecklistRequest represents a request to create a new checklist
type CreateChecklistRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	STIGIDs     []int              `json:"stigIds"`
	TargetData  *models.TargetData `json:"targetData,omitempty"`
	Format      string             `json:"format"` // "ckl" or "cklb"
}

// ChecklistCreationResult represents the result of checklist creation
type ChecklistCreationResult struct {
	ChecklistID int    `json:"checklistId"`
	Name        string `json:"name"`
	Format      string `json:"format"`
	FilePath    string `json:"filePath"`
	RuleCount   int    `json:"ruleCount"`
	STIGCount   int    `json:"stigCount"`
}

// CreateChecklist creates a new checklist from selected STIGs and returns the data structure
func (s *ChecklistService) CreateChecklist(req *CreateChecklistRequest) (*ChecklistCreationResult, error) {
	if len(req.STIGIDs) == 0 {
		return nil, models.ValidationError{
			Field:   "STIGIDs",
			Message: "no STIGs selected",
		}
	}

	// Get STIG rules for all selected STIGs
	totalRules := 0
	var allRules []*models.ImportedSTIGRule
	var stigs []*models.ImportedSTIG

	for _, stigID := range req.STIGIDs {
		stig, err := s.getStigByID(stigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get STIG %d: %w", stigID, err)
		}
		stigs = append(stigs, stig)

		rules, err := s.getStigRulesByStigID(stigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get rules for STIG %d: %w", stigID, err)
		}
		allRules = append(allRules, rules...)
		totalRules += len(rules)
	}

	// Return checklist creation result (frontend will handle file creation and saving)
	return &ChecklistCreationResult{
		Name:      req.Name,
		Format:    req.Format,
		FilePath:  "", // Frontend will determine the actual file path
		RuleCount: totalRules,
		STIGCount: len(req.STIGIDs),
	}, nil
}

// CreateChecklistFromStigs creates a full checklist data structure from selected STIGs
func (s *ChecklistService) CreateChecklistFromStigs(req *CreateChecklistRequest) (*ChecklistWithItems, error) {
	if len(req.STIGIDs) == 0 {
		return nil, models.ValidationError{
			Field:   "STIGIDs",
			Message: "no STIGs selected",
		}
	}

	// Get STIG rules for all selected STIGs
	totalRules := 0
	var allRules []*models.ImportedSTIGRule
	var stigs []*models.ImportedSTIG

	for _, stigID := range req.STIGIDs {
		stig, err := s.getStigByID(stigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get STIG %d: %w", stigID, err)
		}
		stigs = append(stigs, stig)

		rules, err := s.getStigRulesByStigID(stigID)
		if err != nil {
			return nil, fmt.Errorf("failed to get rules for STIG %d: %w", stigID, err)
		}
		allRules = append(allRules, rules...)
		totalRules += len(rules)
	}

	// Create checklist data structure
	checklistData := &ChecklistWithItems{
		Checklist: &models.UserChecklist{
			Name:        req.Name,
			Description: &req.Description,
			TargetData:  nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Items: make([]*ChecklistItemWithRule, 0, totalRules),
		STIGs: stigs,
	}

	// Convert target data if provided
	if req.TargetData != nil {
		targetDataJSON, _ := json.Marshal(req.TargetData)
		targetDataStr := string(targetDataJSON)
		checklistData.Checklist.TargetData = &targetDataStr
	}

	// Create checklist items
	for _, rule := range allRules {
		stig := s.findStigByID(stigs, rule.STIGID)
		if stig == nil {
			continue
		}

		// Determine vulnerability ID
		vulnID := rule.RuleID
		if rule.GroupID != nil && *rule.GroupID != "" &&
			*rule.GroupID != "0" && *rule.GroupID != "N/A" &&
			*rule.GroupID != "NA" && *rule.GroupID != "n/a" &&
			*rule.GroupID != "na" {
			vulnID = *rule.GroupID
		}

		item := &models.ChecklistItem{
			ID:                    len(checklistData.Items) + 1, // Generate unique ID
			VulnID:                vulnID,
			Status:                "NotReviewed",
			FindingDetails:        nil,
			Comments:              nil,
			SeverityOverride:      nil,
			SeverityJustification: nil,
			UpdatedAt:             time.Now(),
		}

		checklistItem := &ChecklistItemWithRule{
			Item: item,
			Rule: rule,
			STIG: stig,
		}

		checklistData.Items = append(checklistData.Items, checklistItem)
	}

	return checklistData, nil
}

// GenerateChecklistFile generates a CKL or CKLb file from checklist data
func (s *ChecklistService) GenerateChecklistFile(checklistData *ChecklistWithItems, format string, outputPath string) error {
	switch format {
	case "ckl":
		return s.generateCKLFile(checklistData, outputPath)
	case "cklb":
		return s.generateCKLbFile(checklistData, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ChecklistWithItems represents a checklist with all its items and metadata
type ChecklistWithItems struct {
	Checklist *models.UserChecklist
	Items     []*ChecklistItemWithRule
	STIGs     []*models.ImportedSTIG
}

// ChecklistItemWithRule combines checklist item with rule details
type ChecklistItemWithRule struct {
	Item *models.ChecklistItem
	Rule *models.ImportedSTIGRule
	STIG *models.ImportedSTIG
}

// generateCKLFile generates a CKL (XML) file
func (s *ChecklistService) generateCKLFile(data *ChecklistWithItems, outputPath string) error {
	// Parse target data
	var targetData *models.TargetData
	if data.Checklist.TargetData != nil && *data.Checklist.TargetData != "" {
		targetData = &models.TargetData{}
		json.Unmarshal([]byte(*data.Checklist.TargetData), targetData)
	}
	if targetData == nil {
		targetData = &models.TargetData{
			TargetType:     "Computing",
			HostName:       "localhost",
			TechnologyArea: "General Purpose Computing",
		}
	}

	// Build CKL structure
	ckl := models.CKL{
		Asset: models.Asset{
			Role:          targetData.Role,
			AssetType:     targetData.TargetType,
			HostName:      targetData.HostName,
			HostIP:        targetData.IPAddress,
			HostMAC:       targetData.MACAddress,
			HostFQDN:      targetData.FQDN,
			TargetComment: targetData.Comments,
			TechArea:      targetData.TechnologyArea,
			WebOrDatabase: "false",
			WebDBSite:     targetData.WebDBSite,
			WebDBInstance: targetData.WebDBInstance,
		},
		STIGs: models.STIGs{},
	}

	// Group items by STIG
	stigGroups := make(map[int][]*ChecklistItemWithRule)
	for _, item := range data.Items {
		stigID := item.STIG.ID
		stigGroups[stigID] = append(stigGroups[stigID], item)
	}

	// Build ISTIGs
	for _, items := range stigGroups {
		stig := items[0].STIG // Get STIG info from first item

		// Build STIG metadata
		stigInfo := models.STIGInfo{
			SIData: []models.SIData{
				{SIDName: "version", SIDData: stig.Version},
				{SIDName: "classification", SIDData: "UNCLASSIFIED"},
				{SIDName: "customname", SIDData: ""},
				{SIDName: "stigid", SIDData: stig.BenchmarkID},
				{SIDName: "description", SIDData: stringOrEmpty(stig.Description)},
				{SIDName: "filename", SIDData: stig.Filename},
				{SIDName: "releaseinfo", SIDData: stringOrEmpty(stig.ReleaseDate)},
				{SIDName: "title", SIDData: stringOrEmpty(stig.Title)},
				{SIDName: "uuid", SIDData: uuid.New().String()},
				{SIDName: "notice", SIDData: "terms-of-use"},
				{SIDName: "source", SIDData: "STIG.DOD.MIL"},
			},
		}

		// Build vulnerabilities
		var vulns []models.Vuln
		for _, item := range items {
			vuln := s.buildCKLVuln(item)
			vulns = append(vulns, vuln)
		}

		istig := models.ISTIG{
			STIGInfo: stigInfo,
			Vulns:    vulns,
		}

		ckl.STIGs.ISTIGs = append(ckl.STIGs.ISTIGs, istig)
	}

	// Generate XML
	output, err := xml.MarshalIndent(ckl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal CKL: %w", err)
	}

	// Write to file
	fullOutput := []byte(xml.Header + string(output))
	return writeToFile(outputPath, fullOutput)
}

// buildCKLVuln builds a CKL vulnerability from checklist item
func (s *ChecklistService) buildCKLVuln(item *ChecklistItemWithRule) models.Vuln {
	rule := item.Rule

	// Parse CCI refs
	var ccis []string
	if rule.CCIRefs != nil {
		json.Unmarshal([]byte(*rule.CCIRefs), &ccis)
	}

	return models.Vuln{
		SVuln: []models.SVuln{
			{VulnAttribute: "Vuln_Num", AttributeData: item.Item.VulnID},
			{VulnAttribute: "Severity", AttributeData: stringOrEmpty(rule.Severity)},
			{VulnAttribute: "Group_Title", AttributeData: stringOrEmpty(rule.GroupTitle)},
			{VulnAttribute: "Rule_ID", AttributeData: rule.RuleID},
			{VulnAttribute: "Rule_Ver", AttributeData: stringOrEmpty(rule.VersionID)},
			{VulnAttribute: "Rule_Title", AttributeData: stringOrEmpty(rule.Title)},
			{VulnAttribute: "Vuln_Discuss", AttributeData: extractVulnDiscussionText(stringOrEmpty(rule.Description))},
			{VulnAttribute: "IA_Controls", AttributeData: ""},
			{VulnAttribute: "Check_Content", AttributeData: stringOrEmpty(rule.CheckContent)},
			{VulnAttribute: "Fix_Text", AttributeData: stringOrEmpty(rule.FixText)},
			{VulnAttribute: "False_Positives", AttributeData: ""},
			{VulnAttribute: "False_Negatives", AttributeData: ""},
			{VulnAttribute: "Documentable", AttributeData: "false"},
			{VulnAttribute: "Mitigations", AttributeData: ""},
			{VulnAttribute: "Potential_Impact", AttributeData: ""},
			{VulnAttribute: "Third_Party_Tools", AttributeData: ""},
			{VulnAttribute: "Mitigation_Control", AttributeData: ""},
			{VulnAttribute: "Responsibility", AttributeData: ""},
			{VulnAttribute: "Security_Override_Guidance", AttributeData: ""},
			{VulnAttribute: "Check_Content_Ref", AttributeData: ""},
			{VulnAttribute: "Weight", AttributeData: floatToString(rule.Weight)},
			{VulnAttribute: "Class", AttributeData: "Unclass"},
			{VulnAttribute: "STIGRef", AttributeData: stringOrEmpty(rule.Title)},
			{VulnAttribute: "TargetKey", AttributeData: ""},
			{VulnAttribute: "STIG_UUID", AttributeData: uuid.New().String()},
			{VulnAttribute: "LEGACY_ID", AttributeData: ""},
			{VulnAttribute: "CCI_REF", AttributeData: joinStrings(ccis, " ")},
		},
		Status:                item.Item.Status,
		FindingDetails:        stringOrEmpty(item.Item.FindingDetails),
		Comments:              stringOrEmpty(item.Item.Comments),
		SeverityOverride:      stringOrEmpty(item.Item.SeverityOverride),
		SeverityJustification: stringOrEmpty(item.Item.SeverityJustification),
	}
}

// Helper functions
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func floatToString(f *float64) string {
	if f == nil {
		return "10.0"
	}
	return fmt.Sprintf("%.1f", *f)
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, str := range strs[1:] {
		result += sep + str
	}
	return result
}

// generateCKLbFile generates a CKLb (JSON) file
func (s *ChecklistService) generateCKLbFile(data *ChecklistWithItems, outputPath string) error {
	// Parse target data
	var targetData *models.TargetData
	if data.Checklist.TargetData != nil && *data.Checklist.TargetData != "" {
		targetData = &models.TargetData{}
		json.Unmarshal([]byte(*data.Checklist.TargetData), targetData)
	}
	if targetData == nil {
		targetData = &models.TargetData{
			TargetType:     "Computing",
			HostName:       "localhost",
			TechnologyArea: "General Purpose Computing",
		}
	}

	// Build CKLb structure
	cklb := models.CKLb{
		Title:       data.Checklist.Name,
		ID:          uuid.New().String(),
		CklbVersion: "1.0",
		TargetData:  targetData,
		STIGs:       []models.STIG{},
	}

	// Group items by STIG
	stigGroups := make(map[int][]*ChecklistItemWithRule)
	for _, item := range data.Items {
		stigID := item.STIG.ID
		stigGroups[stigID] = append(stigGroups[stigID], item)
	}

	// Build STIGs
	for _, items := range stigGroups {
		stigData := items[0].STIG // Get STIG info from first item

		// Build STIG rules
		var rules []models.STIGRule
		for _, item := range items {
			rule := s.buildCKLbRule(item)
			rules = append(rules, rule)
		}

		stig := models.STIG{
			STIGName:            stigData.Name,
			DisplayName:         stringOrEmpty(stigData.Title),
			STIGID:              stigData.BenchmarkID,
			ReleaseInfo:         stringOrEmpty(stigData.ReleaseDate),
			UUID:                uuid.New().String(),
			ReferenceIdentifier: nil,
			Size:                len(rules),
			Version:             stringOrEmpty(&stigData.Version),
			Rules:               rules,
		}

		cklb.STIGs = append(cklb.STIGs, stig)
	}

	// Generate JSON
	output, err := json.MarshalIndent(cklb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal CKLb: %w", err)
	}

	return writeToFile(outputPath, output)
}

// buildCKLbRule builds a CKLb rule from checklist item
func (s *ChecklistService) buildCKLbRule(item *ChecklistItemWithRule) models.STIGRule {
	rule := item.Rule

	// Parse CCI refs
	var ccis []string
	if rule.CCIRefs != nil {
		json.Unmarshal([]byte(*rule.CCIRefs), &ccis)
	}

	return models.STIGRule{
		UUID:                     uuid.New().String(),
		STIGUUID:                 uuid.New().String(),
		TargetKey:                nil,
		STIGRef:                  rule.Title,
		GroupID:                  stringOrEmpty(rule.GroupID),
		GroupIDSrc:               "STIG",
		RuleID:                   rule.RuleID,
		RuleIDSrc:                "STIG",
		Weight:                   floatToString(rule.Weight),
		Classification:           "PUBLIC",
		Severity:                 stringOrEmpty(rule.Severity),
		RuleVersion:              stringOrEmpty(rule.VersionID),
		GroupTitle:               stringOrEmpty(rule.GroupTitle),
		RuleTitle:                stringOrEmpty(rule.Title),
		FixText:                  stringOrEmpty(rule.FixText),
		FalsePositives:           "",
		FalseNegatives:           "",
		Discussion:               extractVulnDiscussionText(stringOrEmpty(rule.Description)),
		CheckContent:             stringOrEmpty(rule.CheckContent),
		Documentable:             "false",
		Mitigations:              "",
		PotentialImpacts:         "",
		ThirdPartyTools:          "",
		MitigationControl:        "",
		Responsibility:           "",
		SecurityOverrideGuidance: "",
		IAControls:               "",
		CheckContentRef:          nil,
		LegacyIDs:                []string{},
		CCIs:                     ccis,
		GroupTree:                []models.GroupTree{},
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		STIGUuid:                 uuid.New().String(),
		Status:                   s.mapFrontendStatusToCKLb(item.Item.Status),
		Overrides:                make(map[string]interface{}),
		Comments:                 stringOrEmpty(item.Item.Comments),
		FindingDetails:           stringOrEmpty(item.Item.FindingDetails),
	}
}

func writeToFile(path string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// getStigByID gets a STIG by ID
func (s *ChecklistService) getStigByID(stigID int) (*models.ImportedSTIG, error) {
	var stig models.ImportedSTIG
	err := s.db.GetConnection().QueryRow(`
		SELECT id, name, version, release_date, benchmark_id, title, description,
		       filename, file_path, file_hash, rule_count, imported_at, updated_at
		FROM stigs WHERE id = ?`, stigID,
	).Scan(
		&stig.ID, &stig.Name, &stig.Version, &stig.ReleaseDate, &stig.BenchmarkID,
		&stig.Title, &stig.Description, &stig.Filename, &stig.FilePath, &stig.FileHash,
		&stig.RuleCount, &stig.ImportedAt, &stig.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &stig, nil
}

// getStigRulesByStigID gets all rules for a STIG
func (s *ChecklistService) getStigRulesByStigID(stigID int) ([]*models.ImportedSTIGRule, error) {
	query := `
		SELECT id, stig_id, rule_id, version_id, title, description, severity, weight,
		       group_id, group_title, check_content, fix_text, cci_refs, created_at
		FROM stig_rules WHERE stig_id = ? ORDER BY rule_id
	`

	rows, err := s.db.GetConnection().Query(query, stigID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.ImportedSTIGRule
	for rows.Next() {
		rule := &models.ImportedSTIGRule{}
		err := rows.Scan(
			&rule.ID, &rule.STIGID, &rule.RuleID, &rule.VersionID, &rule.Title,
			&rule.Description, &rule.Severity, &rule.Weight, &rule.GroupID,
			&rule.GroupTitle, &rule.CheckContent, &rule.FixText, &rule.CCIRefs, &rule.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// findStigByID finds a STIG in a slice by ID
func (s *ChecklistService) findStigByID(stigs []*models.ImportedSTIG, id int) *models.ImportedSTIG {
	for _, stig := range stigs {
		if stig.ID == id {
			return stig
		}
	}
	return nil
}

// LoadChecklistFromFile loads a checklist from a CKL or CKLb file
func (s *ChecklistService) LoadChecklistFromFile(filePath string) (*ChecklistWithItems, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Determine format based on file extension
	format := "cklb"
	if strings.HasSuffix(strings.ToLower(filePath), ".ckl") {
		format = "ckl"
	}

	switch format {
	case "ckl":
		return s.parseCKLFile(data, filePath)
	case "cklb":
		return s.parseCKLbFile(data, filePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}
}

// parseCKLFile parses a CKL (XML) file
func (s *ChecklistService) parseCKLFile(data []byte, filePath string) (*ChecklistWithItems, error) {
	var ckl models.CKL
	if err := xml.Unmarshal(data, &ckl); err != nil {
		return nil, fmt.Errorf("failed to parse CKL file: %w", err)
	}

	// Create checklist data structure
	checklistData := &ChecklistWithItems{
		Checklist: &models.UserChecklist{
			Name:        fmt.Sprintf("Loaded from %s", filepath.Base(filePath)),
			Description: nil, // No description in CKL
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Items: make([]*ChecklistItemWithRule, 0),
		STIGs: make([]*models.ImportedSTIG, 0),
	}

	// Convert target data
	if ckl.Asset.Role != "" || ckl.Asset.HostName != "" {
		targetData := &models.TargetData{
			Role:           ckl.Asset.Role,
			HostName:       ckl.Asset.HostName,
			IPAddress:      ckl.Asset.HostIP,
			MACAddress:     ckl.Asset.HostMAC,
			FQDN:           ckl.Asset.HostFQDN,
			Comments:       ckl.Asset.TargetComment,
			TechnologyArea: ckl.Asset.TechArea,
			WebDBSite:      ckl.Asset.WebDBSite,
			WebDBInstance:  ckl.Asset.WebDBInstance,
		}
		targetDataJSON, _ := json.Marshal(targetData)
		targetDataStr := string(targetDataJSON)
		checklistData.Checklist.TargetData = &targetDataStr
	}

	// Process each iSTIG
	stigMap := make(map[string]*models.ImportedSTIG)
	for _, istig := range ckl.STIGs.ISTIGs {
		// Extract STIG metadata
		stigInfo := make(map[string]string)
		for _, si := range istig.STIGInfo.SIData {
			stigInfo[si.SIDName] = si.SIDData
		}

		stigID := stigInfo["stigid"]
		if stigID == "" {
			continue
		}

		// Create or get STIG
		stig, exists := stigMap[stigID]
		if !exists {
			releaseInfo := stigInfo["releaseinfo"]
			title := stigInfo["title"]
			description := stigInfo["description"]
			stig = &models.ImportedSTIG{
				Name:        title,
				Version:     stigInfo["version"],
				ReleaseDate: &releaseInfo,
				BenchmarkID: stigID,
				Title:       &title,
				Description: &description,
				Filename:    stigInfo["filename"],
			}
			stigMap[stigID] = stig
			checklistData.STIGs = append(checklistData.STIGs, stig)
		}

		// Process rules
		for _, vuln := range istig.Vulns {
			ruleVer := s.extractAttributeData(vuln.SVuln, "Rule_Ver")
			ruleTitle := s.extractAttributeData(vuln.SVuln, "Rule_Title")
			vulnDiscuss := s.extractAttributeData(vuln.SVuln, "Vuln_Discuss")
			severity := s.extractAttributeData(vuln.SVuln, "Severity")
			groupTitle := s.extractAttributeData(vuln.SVuln, "Group_Title")
			checkContent := s.extractAttributeData(vuln.SVuln, "Check_Content")
			fixText := s.extractAttributeData(vuln.SVuln, "Fix_Text")

			rule := &models.ImportedSTIGRule{
				RuleID:       s.extractAttributeData(vuln.SVuln, "Rule_ID"),
				VersionID:    &ruleVer,
				Title:        &ruleTitle,
				Description:  &vulnDiscuss,
				Severity:     &severity,
				GroupID:      &groupTitle,
				GroupTitle:   &groupTitle,
				CheckContent: &checkContent,
				FixText:      &fixText,
			}

			item := &models.ChecklistItem{
				ID:                    len(checklistData.Items) + 1, // Generate unique ID
				VulnID:                s.extractAttributeData(vuln.SVuln, "Vuln_Num"),
				Status:                vuln.Status,
				FindingDetails:        &vuln.FindingDetails,
				Comments:              &vuln.Comments,
				SeverityOverride:      &vuln.SeverityOverride,
				SeverityJustification: &vuln.SeverityJustification,
				UpdatedAt:             time.Now(),
			}

			checklistItem := &ChecklistItemWithRule{
				Item: item,
				Rule: rule,
				STIG: stig,
			}

			checklistData.Items = append(checklistData.Items, checklistItem)
		}
	}

	return checklistData, nil
}

// parseCKLbFile parses a CKLb (JSON) file
func (s *ChecklistService) parseCKLbFile(data []byte, filePath string) (*ChecklistWithItems, error) {
	var cklb models.CKLb
	if err := json.Unmarshal(data, &cklb); err != nil {
		return nil, fmt.Errorf("failed to parse CKLb file: %w", err)
	}

	// Create checklist data structure
	checklistData := &ChecklistWithItems{
		Checklist: &models.UserChecklist{
			Name:        cklb.Title,
			Description: nil, // No description in CKLb
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Items: make([]*ChecklistItemWithRule, 0),
		STIGs: make([]*models.ImportedSTIG, 0),
	}

	// Convert target data
	if cklb.TargetData != nil {
		targetDataJSON, _ := json.Marshal(cklb.TargetData)
		targetDataStr := string(targetDataJSON)
		checklistData.Checklist.TargetData = &targetDataStr
	}

	// Process each STIG
	for _, stigData := range cklb.STIGs {
		stig := &models.ImportedSTIG{
			Name:        stigData.STIGName,
			Version:     stigData.Version,
			ReleaseDate: &stigData.ReleaseInfo,
			BenchmarkID: stigData.STIGID,
			Title:       &stigData.DisplayName,
			Description: &stigData.DisplayName,
		}
		checklistData.STIGs = append(checklistData.STIGs, stig)

		// Process rules
		for _, ruleData := range stigData.Rules {
			rule := &models.ImportedSTIGRule{
				RuleID:       ruleData.RuleID,
				VersionID:    &ruleData.RuleVersion,
				Title:        &ruleData.RuleTitle,
				Description:  &ruleData.Discussion,
				Severity:     &ruleData.Severity,
				GroupID:      &ruleData.GroupID,
				GroupTitle:   &ruleData.GroupTitle,
				CheckContent: &ruleData.CheckContent,
				FixText:      &ruleData.FixText,
			}

			item := &models.ChecklistItem{
				ID:             len(checklistData.Items) + 1, // Generate unique ID
				VulnID:         ruleData.GroupID,
				Status:         s.mapCKLbStatusToFrontend(ruleData.Status),
				FindingDetails: &ruleData.FindingDetails,
				Comments:       &ruleData.Comments,
				UpdatedAt:      ruleData.UpdatedAt,
			}

			checklistItem := &ChecklistItemWithRule{
				Item: item,
				Rule: rule,
				STIG: stig,
			}

			checklistData.Items = append(checklistData.Items, checklistItem)
		}
	}

	return checklistData, nil
}

// extractAttributeData extracts data from SVuln attributes
func (s *ChecklistService) extractAttributeData(svulns []models.SVuln, attributeName string) string {
	for _, svuln := range svulns {
		if svuln.VulnAttribute == attributeName {
			return svuln.AttributeData
		}
	}
	return ""
}

// mapCKLbStatusToFrontend converts CKLb status format to frontend format
func (s *ChecklistService) mapCKLbStatusToFrontend(cklbStatus string) string {
	switch cklbStatus {
	case "not_reviewed":
		return "NotReviewed"
	case "open":
		return "Open"
	case "not_a_finding":
		return "NotAFinding"
	case "not_applicable":
		return "Not_Applicable"
	default:
		return "NotReviewed" // Default to NotReviewed for unknown statuses
	}
}

// mapFrontendStatusToCKLb converts frontend status format to CKLb format
func (s *ChecklistService) mapFrontendStatusToCKLb(frontendStatus string) string {
	switch frontendStatus {
	case "NotReviewed":
		return "not_reviewed"
	case "Open":
		return "open"
	case "NotAFinding":
		return "not_a_finding"
	case "Not_Applicable":
		return "not_applicable"
	default:
		return "not_reviewed" // Default to not_reviewed for unknown statuses
	}
}
