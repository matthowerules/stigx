package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/matthowerules/stigx/internal/converters"
	"github.com/matthowerules/stigx/internal/parsers"
)

var convertCmd = &cobra.Command{
	Use:   "convert [input file] [output file]",
	Short: "Convert between CKL and CKLb formats",
	Long: `Convert checklist files between different formats:
- CKL (XML) to CKLb (JSON)
- CKLb (JSON) to CKL (XML)

The output format is determined by the file extension.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		outputFile := args[1]
		
		validate, _ := cmd.Flags().GetBool("validate")
		force, _ := cmd.Flags().GetBool("force")
		
		fmt.Printf("Converting %s to %s\n", inputFile, outputFile)
		
		err := performConversion(inputFile, outputFile, validate, force)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Conversion completed successfully!")
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
	convertCmd.Flags().BoolP("validate", "v", true, "Validate files against schema")
	convertCmd.Flags().BoolP("force", "f", false, "Overwrite output file if it exists")
}

// performConversion handles the actual conversion logic
func performConversion(inputFile, outputFile string, validate, force bool) error {
	// Check if input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}
	
	// Check if output file exists and force flag
	if _, err := os.Stat(outputFile); err == nil && !force {
		return fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputFile)
	}
	
	// Detect input format
	detector := parsers.NewFileDetector()
	inputType, err := detector.DetectFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to detect input file format: %w", err)
	}
	
	fmt.Printf("Detected input format: %s\n", inputType.String())
	
	// Determine output format
	outputType, err := detector.SuggestOutputFormat(inputFile, outputFile)
	if err != nil {
		return fmt.Errorf("failed to determine output format: %w", err)
	}
	
	fmt.Printf("Target output format: %s\n", outputType.String())
	
	// Validate input file if requested
	if validate {
		fmt.Println("Validating input file...")
		if err := validateFile(inputFile, inputType); err != nil {
			return fmt.Errorf("input file validation failed: %w", err)
		}
		fmt.Println("Input file validation passed!")
	}
	
	// Perform conversion
	converter := converters.NewConverter()
	if err := converter.ConvertFileWithFormats(inputFile, outputFile, inputType, outputType); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}
	
	// Validate output file if requested
	if validate {
		fmt.Println("Validating output file...")
		if err := validateFile(outputFile, outputType); err != nil {
			return fmt.Errorf("output file validation failed: %w", err)
		}
		fmt.Println("Output file validation passed!")
	}
	
	// Show conversion statistics
	if err := showConversionStats(inputFile, outputFile, inputType, outputType); err != nil {
		fmt.Printf("Warning: failed to show conversion statistics: %v\n", err)
	}
	
	return nil
}

// validateFile validates a file against its expected format
func validateFile(filePath string, fileType parsers.FileType) error {
	switch fileType {
	case parsers.FileTypeCKL:
		parser := parsers.NewCKLParser()
		_, err := parser.ParseFile(filePath)
		return err
	case parsers.FileTypeCKLb:
		parser := parsers.NewCKLbParser()
		_, err := parser.ParseFile(filePath)
		return err
	default:
		return fmt.Errorf("validation not supported for file type: %s", fileType.String())
	}
}

// showConversionStats displays statistics about the conversion
func showConversionStats(inputFile, outputFile string, inputType, outputType parsers.FileType) error {
	detector := parsers.NewFileDetector()
	
	// Get input file stats
	inputStats, err := detector.GetFileInfo(inputFile)
	if err != nil {
		return fmt.Errorf("failed to get input file stats: %w", err)
	}
	
	// Get output file stats
	outputStats, err := detector.GetFileInfo(outputFile)
	if err != nil {
		return fmt.Errorf("failed to get output file stats: %w", err)
	}
	
	fmt.Println("\n=== Conversion Statistics ===")
	fmt.Printf("Input:  %s (%s) - %v bytes\n", 
		inputFile, inputType.String(), inputStats["file_size"])
	fmt.Printf("Output: %s (%s) - %v bytes\n", 
		outputFile, outputType.String(), outputStats["file_size"])
	
	// Show detailed stats if available
	if inputStatistics, ok := inputStats["statistics"].(map[string]interface{}); ok {
		if totalRules, exists := inputStatistics["total_rules"]; exists {
			fmt.Printf("Total rules/vulnerabilities: %v\n", totalRules)
		} else if totalVulns, exists := inputStatistics["total_vulnerabilities"]; exists {
			fmt.Printf("Total rules/vulnerabilities: %v\n", totalVulns)
		}
		
		if stigCount, exists := inputStatistics["stig_count"]; exists {
			fmt.Printf("Number of STIGs: %v\n", stigCount)
		}
	}
	
	return nil
}
