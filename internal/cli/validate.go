package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/matthowerules/stigx/internal/models"
	"github.com/matthowerules/stigx/internal/parsers"
)

var validateCmd = &cobra.Command{
	Use:   "validate [file...]",
	Short: "Validate CKL/CKLb files against their schemas",
	Long: `Validate one or more checklist files against their official schemas:
- CKL files are validated against XSD
- CKLb files are validated against JSON Schema
- Reports any schema violations or format issues`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		strict, _ := cmd.Flags().GetBool("strict")
		quiet, _ := cmd.Flags().GetBool("quiet")
		
		hasErrors := false
		
		for _, file := range args {
			if !quiet {
				fmt.Printf("Validating %s...\n", file)
			}
			
			err := performValidation(file, strict, quiet)
			if err != nil {
				fmt.Printf("ERROR: %s: %v\n", file, err)
				hasErrors = true
			} else if !quiet {
				fmt.Printf("✓ %s: Valid\n", file)
			}
		}
		
		if hasErrors {
			os.Exit(1)
		} else if !quiet {
			fmt.Printf("\nAll %d file(s) passed validation!\n", len(args))
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolP("strict", "s", false, "Enable strict validation mode")
	validateCmd.Flags().BoolP("quiet", "q", false, "Only show errors")
}

// performValidation validates a single file
func performValidation(filePath string, strict, quiet bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist")
	}
	
	// Detect file format
	detector := parsers.NewFileDetector()
	fileType, err := detector.DetectFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to detect file format: %w", err)
	}
	
	if !quiet {
		fmt.Printf("  Format: %s\n", fileType.String())
	}
	
	// Validate based on file type
	switch fileType {
	case parsers.FileTypeCKL:
		return validateCKLFile(filePath, strict, quiet)
	case parsers.FileTypeCKLb:
		return validateCKLbFile(filePath, strict, quiet)
	case parsers.FileTypeUnknown:
		return fmt.Errorf("unknown or unsupported file format")
	default:
		return fmt.Errorf("validation not yet supported for %s format", fileType.String())
	}
}

// validateCKLFile validates a CKL file
func validateCKLFile(filePath string, strict, quiet bool) error {
	parser := parsers.NewCKLParser()
	
	// Parse the file
	ckl, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	
	// Get and display statistics
	if !quiet {
		stats := parser.GetStatistics(ckl)
		fmt.Printf("  Asset: %v (%v)\n", stats["host_name"], stats["asset_type"])
		fmt.Printf("  STIGs: %v\n", stats["stig_count"])
		fmt.Printf("  Vulnerabilities: %v\n", stats["total_vulnerabilities"])
		
		if statusCounts, ok := stats["status_counts"].(map[string]int); ok && len(statusCounts) > 0 {
			fmt.Print("  Status breakdown: ")
			for status, count := range statusCounts {
				fmt.Printf("%s=%d ", status, count)
			}
			fmt.Println()
		}
	}
	
	// Additional strict validation
	if strict {
		return performStrictCKLValidation(ckl, quiet)
	}
	
	return nil
}

// validateCKLbFile validates a CKLb file
func validateCKLbFile(filePath string, strict, quiet bool) error {
	parser := parsers.NewCKLbParser()
	
	// Parse the file
	cklb, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}
	
	// Get and display statistics
	if !quiet {
		stats := parser.GetStatistics(cklb)
		fmt.Printf("  Title: %v\n", stats["title"])
		fmt.Printf("  STIGs: %v\n", stats["stig_count"])
		fmt.Printf("  Rules: %v\n", stats["total_rules"])
		
		if targetHost := stats["target_hostname"]; targetHost != nil && targetHost != "" {
			fmt.Printf("  Target: %v (%v)\n", targetHost, stats["target_ip"])
		}
		
		if statusCounts, ok := stats["status_counts"].(map[string]int); ok && len(statusCounts) > 0 {
			fmt.Print("  Status breakdown: ")
			for status, count := range statusCounts {
				fmt.Printf("%s=%d ", status, count)
			}
			fmt.Println()
		}
	}
	
	// Additional strict validation
	if strict {
		return performStrictCKLbValidation(cklb, quiet)
	}
	
	return nil
}

// performStrictCKLValidation performs additional strict validation for CKL files
func performStrictCKLValidation(ckl *models.CKL, quiet bool) error {
	var warnings []string
	
	// Check for missing asset information
	if ckl.Asset.HostName == "" {
		warnings = append(warnings, "host name is empty")
	}
	if ckl.Asset.HostIP == "" {
		warnings = append(warnings, "host IP is empty")
	}
	
	// Check for vulnerabilities without findings
	unreviewed := 0
	for _, istig := range ckl.STIGs.ISTIGs {
		for _, vuln := range istig.Vulns {
			if vuln.Status == "" || vuln.Status == "Not_Reviewed" {
				unreviewed++
			}
		}
	}
	
	if unreviewed > 0 {
		warnings = append(warnings, fmt.Sprintf("%d vulnerabilities are not reviewed", unreviewed))
	}
	
	// Report warnings
	if len(warnings) > 0 && !quiet {
		fmt.Printf("  Warnings: %s\n", fmt.Sprintf("%v", warnings))
	}
	
	return nil
}

// performStrictCKLbValidation performs additional strict validation for CKLb files
func performStrictCKLbValidation(cklb *models.CKLb, quiet bool) error {
	var warnings []string
	
	// Check for missing target data
	if cklb.TargetData == nil {
		warnings = append(warnings, "target data is missing")
	} else {
		if cklb.TargetData.HostName == "" {
			warnings = append(warnings, "host name is empty")
		}
		if cklb.TargetData.IPAddress == "" {
			warnings = append(warnings, "IP address is empty")
		}
	}
	
	// Check for rules without findings
	unreviewed := 0
	for _, stig := range cklb.STIGs {
		for _, rule := range stig.Rules {
			if rule.Status == "" || rule.Status == "not_reviewed" {
				unreviewed++
			}
		}
	}
	
	if unreviewed > 0 {
		warnings = append(warnings, fmt.Sprintf("%d rules are not reviewed", unreviewed))
	}
	
	// Report warnings
	if len(warnings) > 0 && !quiet {
		fmt.Printf("  Warnings: %s\n", fmt.Sprintf("%v", warnings))
	}
	
	return nil
}
