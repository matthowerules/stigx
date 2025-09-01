package parsers

import (
	"strings"
	"testing"

	"github.com/matthowerules/stigx/internal/models"
)

func TestCKLbParser_Parse(t *testing.T) {
	parser := NewCKLbParser()

	t.Run("Valid CKLb JSON", func(t *testing.T) {
		validJSON := `{
			"title": "Test Checklist",
			"id": "12345678-1234-1234-1234-123456789012",
			"cklb_version": "1.0",
			"stigs": [
				{
					"stig_name": "Test STIG",
					"display_name": "Test STIG Display",
					"stig_id": "TEST_STIG",
					"release_info": "Release: 1 Benchmark Date: 01 Jan 2024",
					"uuid": "87654321-4321-4321-4321-210987654321",
					"size": 1,
					"rules": [
						{
							"uuid": "11111111-1111-1111-1111-111111111111",
							"stig_uuid": "87654321-4321-4321-4321-210987654321",
							"group_id": "V-123456",
							"rule_id": "SV-123456r1",
							"severity": "medium",
							"status": "not_reviewed",
							"comments": "",
							"finding_details": "",
							"createdAt": "2024-01-01T00:00:00.000Z",
							"updatedAt": "2024-01-01T00:00:00.000Z",
							"overrides": {}
						}
					]
				}
			]
		}`

		reader := strings.NewReader(validJSON)
		cklb, err := parser.Parse(reader)

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if cklb.Title != "Test Checklist" {
			t.Errorf("Expected title 'Test Checklist', got: %s", cklb.Title)
		}

		if len(cklb.STIGs) != 1 {
			t.Errorf("Expected 1 STIG, got: %d", len(cklb.STIGs))
		}

		if len(cklb.STIGs[0].Rules) != 1 {
			t.Errorf("Expected 1 rule, got: %d", len(cklb.STIGs[0].Rules))
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{"title": "Test", "invalid": json}`
		reader := strings.NewReader(invalidJSON)

		_, err := parser.Parse(reader)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		missingFieldsJSON := `{"title": ""}`
		reader := strings.NewReader(missingFieldsJSON)

		_, err := parser.Parse(reader)
		if err == nil {
			t.Error("Expected error for missing required fields, got nil")
		}

		if !strings.Contains(err.Error(), "missing required field") {
			t.Errorf("Expected 'missing required field' error, got: %v", err)
		}
	})

	t.Run("Invalid UUID Format", func(t *testing.T) {
		invalidUUIDJSON := `{
			"title": "Test",
			"id": "invalid-uuid-format"
		}`
		reader := strings.NewReader(invalidUUIDJSON)

		_, err := parser.Parse(reader)
		if err == nil {
			t.Error("Expected error for invalid UUID, got nil")
		}

		if !strings.Contains(err.Error(), "invalid UUID format") {
			t.Errorf("Expected 'invalid UUID format' error, got: %v", err)
		}
	})
}

func TestCKLbParser_ValidateBasic(t *testing.T) {
	parser := NewCKLbParser()

	t.Run("Valid CKLb", func(t *testing.T) {
		cklb := &models.CKLb{
			Title: "Test Checklist",
			ID:    "12345678-1234-1234-1234-123456789012",
		}

		err := parser.ValidateBasic(cklb)
		if err != nil {
			t.Errorf("Expected no error for valid CKLb, got: %v", err)
		}
	})

	t.Run("Missing Title", func(t *testing.T) {
		cklb := &models.CKLb{
			ID: "12345678-1234-1234-1234-123456789012",
		}

		err := parser.ValidateBasic(cklb)
		if err == nil {
			t.Error("Expected error for missing title, got nil")
		}
	})

	t.Run("Missing ID", func(t *testing.T) {
		cklb := &models.CKLb{
			Title: "Test Checklist",
		}

		err := parser.ValidateBasic(cklb)
		if err == nil {
			t.Error("Expected error for missing ID, got nil")
		}
	})
}

func TestCKLbParser_IsCKLbFile(t *testing.T) {
	parser := NewCKLbParser()

	testCases := []struct {
		filePath string
		expected bool
	}{
		{"test.cklb", true},
		{"test.json", true},
		{"test.CKLB", true},
		{"test.JSON", true},
		{"test.ckl", false},
		{"test.xml", false},
		{"test.txt", false},
		{"test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filePath, func(t *testing.T) {
			result := parser.IsCKLbFile(tc.filePath)
			if result != tc.expected {
				t.Errorf("Expected %v for %s, got %v", tc.expected, tc.filePath, result)
			}
		})
	}
}

func TestCKLbParser_GetStatistics(t *testing.T) {
	parser := NewCKLbParser()

	cklb := &models.CKLb{
		Title: "Test Checklist",
		ID:    "12345678-1234-1234-1234-123456789012",
		TargetData: &models.TargetData{
			HostName:  "test-host",
			IPAddress: "192.168.1.1",
		},
		STIGs: []models.STIG{
			{
				STIGName: "Test STIG",
				Rules: []models.STIGRule{
					{
						Status:   "open",
						Severity: "high",
					},
					{
						Status:   "not_a_finding",
						Severity: "medium",
					},
				},
			},
		},
	}

	stats := parser.GetStatistics(cklb)

	if stats["title"] != "Test Checklist" {
		t.Errorf("Expected title 'Test Checklist', got: %v", stats["title"])
	}

	if stats["stig_count"] != 1 {
		t.Errorf("Expected stig_count 1, got: %v", stats["stig_count"])
	}

	if stats["total_rules"] != 2 {
		t.Errorf("Expected total_rules 2, got: %v", stats["total_rules"])
	}

	if stats["target_hostname"] != "test-host" {
		t.Errorf("Expected target_hostname 'test-host', got: %v", stats["target_hostname"])
	}

	statusCounts, ok := stats["status_counts"].(map[string]int)
	if !ok {
		t.Error("Expected status_counts to be map[string]int")
	} else {
		if statusCounts["open"] != 1 {
			t.Errorf("Expected 1 open status, got: %d", statusCounts["open"])
		}
		if statusCounts["not_a_finding"] != 1 {
			t.Errorf("Expected 1 not_a_finding status, got: %d", statusCounts["not_a_finding"])
		}
	}

	severityCounts, ok := stats["severity_counts"].(map[string]int)
	if !ok {
		t.Error("Expected severity_counts to be map[string]int")
	} else {
		if severityCounts["high"] != 1 {
			t.Errorf("Expected 1 high severity, got: %d", severityCounts["high"])
		}
		if severityCounts["medium"] != 1 {
			t.Errorf("Expected 1 medium severity, got: %d", severityCounts["medium"])
		}
	}
}
