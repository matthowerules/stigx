package models

import "encoding/xml"

// CKL represents the legacy STIG Viewer Checklist format (XML)
type CKL struct {
	XMLName xml.Name `xml:"CHECKLIST"`
	Asset   Asset    `xml:"ASSET"`
	STIGs   STIGs    `xml:"STIGS"`
}

// Asset represents the target system information
type Asset struct {
	Role             string `xml:"ROLE"`
	AssetType        string `xml:"ASSET_TYPE"`
	Marking          string `xml:"MARKING"`
	HostName         string `xml:"HOST_NAME"`
	HostIP           string `xml:"HOST_IP"`
	HostMAC          string `xml:"HOST_MAC"`
	HostFQDN         string `xml:"HOST_FQDN"`
	TargetComment    string `xml:"TARGET_COMMENT"`
	TechArea         string `xml:"TECH_AREA"`
	TargetKey        string `xml:"TARGET_KEY"`
	WebOrDatabase    string `xml:"WEB_OR_DATABASE"`
	WebDBSite        string `xml:"WEB_DB_SITE"`
	WebDBInstance    string `xml:"WEB_DB_INSTANCE"`
}

// STIGs contains the STIG information and rules
type STIGs struct {
	ISTIGs []ISTIG `xml:"iSTIG"`
}

// ISTIG represents an individual STIG
type ISTIG struct {
	STIGInfo STIGInfo   `xml:"STIG_INFO"`
	Vulns    []Vuln     `xml:"VULN"`
}

// STIGInfo contains metadata about the STIG
type STIGInfo struct {
	SIData []SIData `xml:"SI_DATA"`
}

// SIData represents key-value pairs for STIG information
type SIData struct {
	SIDName string `xml:"SID_NAME"`
	SIDData string `xml:"SID_DATA"`
}

// Vuln represents a vulnerability/rule in the STIG
type Vuln struct {
	SVuln []SVuln `xml:"STIG_DATA"`
	Status string `xml:"STATUS"`
	FindingDetails string `xml:"FINDING_DETAILS"`
	Comments string `xml:"COMMENTS"`
	SeverityOverride string `xml:"SEVERITY_OVERRIDE"`
	SeverityJustification string `xml:"SEVERITY_JUSTIFICATION"`
}

// SVuln represents vulnerability data
type SVuln struct {
	VulnAttribute string `xml:"VULN_ATTRIBUTE"`
	AttributeData string `xml:"ATTRIBUTE_DATA"`
}

// Helper structs for easier access to commonly used STIG info
type STIGMetadata struct {
	Version         string
	Classification  string
	CustomName      string
	STIGID          string
	Description     string
	Filename        string
	ReleaseInfo     string
	Title           string
	UUID            string
	Notice          string
	Source          string
}

// Helper structs for easier access to vulnerability data
type VulnMetadata struct {
	VulnNum               string
	Severity              string
	GroupTitle            string
	RuleID                string
	RuleVer               string
	RuleTitle             string
	VulnDiscuss           string
	IAControls            string
	CheckContent          string
	FixText               string
	FalsePositives        string
	FalseNegatives        string
	Documentable          string
	Mitigations           string
	PotentialImpacts      string
	ThirdPartyTools       string
	MitigationControl     string
	Responsibility        string
	SecurityOverrideGuidance string
	CheckContentRef       string
	Weight                string
	Class                 string
	STIGRef               string
	TargetKey             string
	STIGUuid              string
	LegacyID              []string
	CCI                   []string
}