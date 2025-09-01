package models

import (
	"time"
)

// CKLb represents the STIG Viewer 3 Checklist format (JSON)
type CKLb struct {
	Title       string      `json:"title"`
	ID          string      `json:"id"`
	CklbVersion string      `json:"cklb_version,omitempty"`
	Active      *bool       `json:"active,omitempty"`
	Mode        *int        `json:"mode,omitempty"`
	HasPath     *bool       `json:"has_path,omitempty"`
	TargetData  *TargetData `json:"target_data,omitempty"`
	STIGs       []STIG      `json:"stigs,omitempty"`
}

// TargetData represents properties of the scanned system
type TargetData struct {
	TargetType        string `json:"target_type,omitempty"`
	HostName          string `json:"host_name,omitempty"`
	IPAddress         string `json:"ip_address,omitempty"`
	MACAddress        string `json:"mac_address,omitempty"`
	FQDN              string `json:"fqdn,omitempty"`
	Comments          string `json:"comments,omitempty"`
	Role              string `json:"role,omitempty"`
	IsWebDatabase     *bool  `json:"is_web_database,omitempty"`
	TechnologyArea    string `json:"technology_area,omitempty"`
	WebDBSite         string `json:"web_db_site,omitempty"`
	WebDBInstance     string `json:"web_db_instance,omitempty"`
}

// STIG represents a Security Technical Implementation Guide
type STIG struct {
	STIGName            string     `json:"stig_name"`
	DisplayName         string     `json:"display_name"`
	STIGID              string     `json:"stig_id"`
	ReleaseInfo         string     `json:"release_info"`
	UUID                string     `json:"uuid"`
	ReferenceIdentifier *string    `json:"reference_identifier,omitempty"`
	Size                int        `json:"size"`
	Version             string     `json:"version,omitempty"`
	Rules               []STIGRule `json:"rules,omitempty"`
}

// STIGRule represents a rule from a STIG with finding data
type STIGRule struct {
	UUID                        string      `json:"uuid"`
	STIGUUID                    string      `json:"stig_uuid"`
	TargetKey                   *string     `json:"target_key,omitempty"`
	STIGRef                     *string     `json:"stig_ref,omitempty"`
	GroupID                     string      `json:"group_id"`
	GroupIDSrc                  string      `json:"group_id_src"`
	RuleID                      string      `json:"rule_id"`
	RuleIDSrc                   string      `json:"rule_id_src"`
	Weight                      string      `json:"weight"`
	Classification              string      `json:"classification"`
	Severity                    string      `json:"severity"`
	RuleVersion                 string      `json:"rule_version"`
	GroupTitle                  string      `json:"group_title"`
	RuleTitle                   string      `json:"rule_title"`
	FixText                     string      `json:"fix_text"`
	FalsePositives              string      `json:"false_positives"`
	FalseNegatives              string      `json:"false_negatives"`
	Discussion                  string      `json:"discussion"`
	CheckContent                string      `json:"check_content"`
	Documentable                string      `json:"documentable"`
	Mitigations                 string      `json:"mitigations"`
	PotentialImpacts            string      `json:"potential_impacts"`
	ThirdPartyTools             string      `json:"third_party_tools"`
	MitigationControl           string      `json:"mitigation_control"`
	Responsibility              string      `json:"responsibility"`
	SecurityOverrideGuidance    string      `json:"security_override_guidance"`
	IAControls                  string      `json:"ia_controls"`
	CheckContentRef             *ContentRef `json:"check_content_ref,omitempty"`
	LegacyIDs                   []string    `json:"legacy_ids,omitempty"`
	CCIs                        []string    `json:"ccis,omitempty"`
	GroupTree                   []GroupTree `json:"group_tree,omitempty"`
	CreatedAt                   time.Time   `json:"createdAt"`
	UpdatedAt                   time.Time   `json:"updatedAt"`
	STIGUuid                    string      `json:"STIGUuid"` // Note: different casing from stig_uuid
	
	// Finding status and assessment data
	Status          string                 `json:"status"`           // not_reviewed, open, not_a_finding, etc.
	Overrides       map[string]interface{} `json:"overrides"`
	Comments        string                 `json:"comments"`
	FindingDetails  string                 `json:"finding_details"`
}

// ContentRef represents a reference to content
type ContentRef struct {
	Href string `json:"href"`
	Name string `json:"name"`
}

// GroupTree represents hierarchical grouping information
type GroupTree struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}