package models

import (
	"time"
)

// ImportedSTIG represents a Security Technical Implementation Guide in the database
type ImportedSTIG struct {
	ID           int       `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Version      string    `json:"version" db:"version"`
	ReleaseDate  *string   `json:"releaseDate" db:"release_date"`
	BenchmarkID  string    `json:"benchmarkId" db:"benchmark_id"`
	Title        *string   `json:"title" db:"title"`
	Description  *string   `json:"description" db:"description"`
	Filename     string    `json:"filename" db:"filename"`
	FilePath     *string   `json:"filePath" db:"file_path"`
	FileHash     *string   `json:"fileHash" db:"file_hash"`
	RuleCount    int       `json:"ruleCount" db:"rule_count"`
	ImportedAt   time.Time `json:"importedAt" db:"imported_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// ImportedSTIGRule represents an individual rule within a STIG
type ImportedSTIGRule struct {
	ID          int      `json:"id" db:"id"`
	STIGID      int      `json:"stigId" db:"stig_id"`
	RuleID      string   `json:"ruleId" db:"rule_id"`
	VersionID   *string  `json:"versionId" db:"version_id"`
	Title       *string  `json:"title" db:"title"`
	Description *string  `json:"description" db:"description"`
	Severity    *string  `json:"severity" db:"severity"`
	Weight      *float64 `json:"weight" db:"weight"`
	GroupID     *string  `json:"groupId" db:"group_id"`
	GroupTitle  *string  `json:"groupTitle" db:"group_title"`
	CheckContent *string `json:"checkContent" db:"check_content"`
	FixText     *string  `json:"fixText" db:"fix_text"`
	CCIRefs     *string  `json:"cciRefs" db:"cci_refs"` // JSON array as string
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// Bundle represents an imported STIG bundle file
type Bundle struct {
	ID         int       `json:"id" db:"id"`
	Filename   string    `json:"filename" db:"filename"`
	FilePath   string    `json:"filePath" db:"file_path"`
	FileHash   string    `json:"fileHash" db:"file_hash"`
	STIGCount  int       `json:"stigCount" db:"stig_count"`
	ImportedAt time.Time `json:"importedAt" db:"imported_at"`
}

// UserChecklist represents a user-created checklist
type UserChecklist struct {
	ID              int       `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description" db:"description"`
	TargetData      *string   `json:"targetData" db:"target_data"` // JSON
	CreatedFromSTIGs string   `json:"createdFromStigs" db:"created_from_stigs"` // JSON array
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time `json:"updatedAt" db:"updated_at"`
}

// ChecklistItem represents an individual rule instance in a checklist
type ChecklistItem struct {
	ID                     int       `json:"id" db:"id"`
	ChecklistID            int       `json:"checklistId" db:"checklist_id"`
	STIGRuleID             int       `json:"stigRuleId" db:"stig_rule_id"`
	VulnID                 string    `json:"vulnId" db:"vuln_id"`
	Status                 string    `json:"status" db:"status"`
	FindingDetails         *string   `json:"findingDetails" db:"finding_details"`
	Comments               *string   `json:"comments" db:"comments"`
	SeverityOverride       *string   `json:"severityOverride" db:"severity_override"`
	SeverityJustification  *string   `json:"severityJustification" db:"severity_justification"`
	UpdatedAt              time.Time `json:"updatedAt" db:"updated_at"`
}

// ImportedSTIGSummary represents summary information about a STIG
type ImportedSTIGSummary struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Version     string    `json:"version" db:"version"`
	ReleaseDate *string   `json:"releaseDate" db:"release_date"`
	Title       *string   `json:"title" db:"title"`
	RuleCount   int       `json:"ruleCount" db:"rule_count"`
	ImportedAt  time.Time `json:"importedAt" db:"imported_at"`
}

// CreateChecklistRequest represents the request to create a new checklist
type CreateChecklistRequest struct {
	Name         string   `json:"name"`
	Description  *string  `json:"description"`
	STIGIDs      []int    `json:"stigIds"`
	TargetData   *string  `json:"targetData"`
}

// ImportProgress represents progress during STIG bundle import
type ImportProgress struct {
	Stage           string `json:"stage"`           // "extracting", "parsing", "storing"
	CurrentFile     string `json:"currentFile"`     // Current file being processed
	ProcessedFiles  int    `json:"processedFiles"`  // Files processed so far
	TotalFiles      int    `json:"totalFiles"`      // Total files to process
	ProcessedSTIGs  int    `json:"processedStigs"`  // STIGs processed so far
	TotalSTIGs      int    `json:"totalStigs"`      // Total STIGs found
	Message         string `json:"message"`         // Status message
	Error           string `json:"error,omitempty"` // Error message if any
}