package services

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/matthowerules/stigx/internal/database"
	"github.com/matthowerules/stigx/internal/models"
)

// STIGService handles STIG-related operations
type STIGService struct {
	db *database.DB
}

// NewSTIGService creates a new STIG service
func NewSTIGService(db *database.DB) *STIGService {
	return &STIGService{db: db}
}

// parseVersionAndReleaseFromFilename extracts version and release from STIG filename
// Example: "U_Active_Directory_Domain_V3R5_STIG" -> Version: "3", Release: "5"
func parseVersionAndReleaseFromFilename(filename string) (version string, release string) {
	// Regex to match V<number>R<number> pattern in filename
	re := regexp.MustCompile(`V(\d+)R(\d+)`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) >= 3 {
		return matches[1], matches[2]
	}

	// Fallback: try to match just V<number> pattern
	versionRe := regexp.MustCompile(`V(\d+)`)
	versionMatches := versionRe.FindStringSubmatch(filename)
	if len(versionMatches) >= 2 {
		return versionMatches[1], ""
	}

	return "", ""
}

// extractReleaseFromPlainText extracts release number from XCCDF plain-text elements
// Example: "Release: 5 Benchmark Date: 13 Sep 2024" -> "5"
func extractReleaseFromPlainText(plainTexts []XCCDFPlainText) string {
	for _, pt := range plainTexts {
		if pt.ID == "release-info" {
			// Extract release number from format "Release: X ..."
			re := regexp.MustCompile(`Release:\s*(\d+)`)
			matches := re.FindStringSubmatch(pt.Text)
			if len(matches) >= 2 {
				return matches[1]
			}
		}
	}
	return ""
}

// GetAllSTIGs returns all imported STIGs
func (s *STIGService) GetAllSTIGs() ([]*models.ImportedSTIGSummary, error) {
	query := `SELECT id, name, version, release_date, title, rule_count, imported_at FROM stig_summary ORDER BY name, version`

	rows, err := s.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query STIGs: %w", err)
	}
	defer rows.Close()

	var stigs []*models.ImportedSTIGSummary
	for rows.Next() {
		stig := &models.ImportedSTIGSummary{}
		err := rows.Scan(
			&stig.ID, &stig.Name, &stig.Version, &stig.ReleaseDate,
			&stig.Title, &stig.RuleCount, &stig.ImportedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan STIG row: %w", err)
		}
		stigs = append(stigs, stig)
	}

	return stigs, nil
}

// DeleteSTIG deletes a STIG and all its associated rules by ID
func (s *STIGService) DeleteSTIG(id int) error {
	// Start transaction
	tx, err := s.db.GetConnection().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete STIG (this will cascade delete all associated rules due to foreign key constraint)
	result, err := tx.Exec("DELETE FROM stigs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete STIG: %w", err)
	}

	// Check if any row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("STIG with ID %d not found", id)
	}

	// Clean up bundle records for bundles that no longer have any STIGs
	if err := s.cleanupEmptyBundles(tx); err != nil {
		return fmt.Errorf("failed to cleanup empty bundles: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteSTIGs deletes multiple STIGs by their IDs
func (s *STIGService) DeleteSTIGs(ids []int) error {
	if len(ids) == 0 {
		return nil // Nothing to delete
	}

	// Start transaction
	tx, err := s.db.GetConnection().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	// Delete STIGs (this will cascade delete all associated rules due to foreign key constraint)
	query := fmt.Sprintf("DELETE FROM stigs WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete STIGs: %w", err)
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no STIGs found with the provided IDs")
	}

	// Clean up bundle records for bundles that no longer have any STIGs
	if err := s.cleanupEmptyBundles(tx); err != nil {
		return fmt.Errorf("failed to cleanup empty bundles: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSTIG returns a STIG by ID
func (s *STIGService) GetSTIG(id int) (*models.ImportedSTIG, error) {
	query := `SELECT id, name, version, release_date, benchmark_id, title, description, filename, file_path, file_hash, rule_count, imported_at, updated_at FROM stigs WHERE id = ?`

	stig := &models.ImportedSTIG{}
	err := s.db.GetConnection().QueryRow(query, id).Scan(
		&stig.ID, &stig.Name, &stig.Version, &stig.ReleaseDate, &stig.BenchmarkID,
		&stig.Title, &stig.Description, &stig.Filename, &stig.FilePath,
		&stig.FileHash, &stig.RuleCount, &stig.ImportedAt, &stig.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("STIG not found")
		}
		return nil, fmt.Errorf("failed to get STIG: %w", err)
	}

	return stig, nil
}

// ImportBundle imports a STIG bundle zip file
func (s *STIGService) ImportBundle(bundlePath string, progressCallback func(*models.ImportProgress)) error {
	// Calculate file hash
	hash, err := calculateFileHash(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to calculate bundle hash: %w", err)
	}

	filename := filepath.Base(bundlePath)

	// Check if bundle already imported
	exists, err := s.bundleExists(hash)
	if err != nil {
		return fmt.Errorf("failed to check bundle existence: %w", err)
	}
	if exists {
		// Check if there are any STIGs from this bundle still in the database
		stigsExist, err := s.bundleHasSTIGs(filename)
		if err != nil {
			return fmt.Errorf("failed to check if bundle has STIGs: %w", err)
		}
		if stigsExist {
			return fmt.Errorf("bundle already imported")
		}
		// If bundle exists but has no STIGs, we can re-import (delete old bundle record)
		if err := s.deleteBundleRecord(hash); err != nil {
			return fmt.Errorf("failed to clean up old bundle record: %w", err)
		}
	}

	if progressCallback != nil {
		progressCallback(&models.ImportProgress{
			Stage:   "extracting",
			Message: fmt.Sprintf("Opening bundle: %s", filename),
		})
	}

	// Open bundle zip file
	bundle, err := zip.OpenReader(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to open bundle: %w", err)
	}
	defer bundle.Close()

	// Start transaction
	tx, err := s.db.GetConnection().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert bundle record
	_, err = tx.Exec(`
		INSERT INTO bundles (filename, file_path, file_hash)
		VALUES (?, ?, ?)`,
		filename, bundlePath, hash,
	)
	if err != nil {
		return fmt.Errorf("failed to insert bundle: %w", err)
	}

	// Process XCCDF files
	totalFiles := 0
	processedFiles := 0

	// Debug: List all files in bundle
	fmt.Printf("DEBUG: Bundle contains %d files:\n", len(bundle.File))
	for i, file := range bundle.File {
		fmt.Printf("DEBUG: File[%d]: %s\n", i, file.Name)
		if strings.HasSuffix(strings.ToLower(file.Name), "xccdf.xml") {
			fmt.Printf("DEBUG: Found XCCDF file: %s\n", file.Name)
			totalFiles++
		} else if strings.HasSuffix(strings.ToLower(file.Name), ".zip") && !file.FileInfo().IsDir() {
			fmt.Printf("DEBUG: Found nested ZIP file: %s\n", file.Name)
			// Count XCCDF files in nested ZIP
			nestedCount, err := s.countXCCDFInNestedZip(file)
			if err != nil {
				fmt.Printf("DEBUG: Error checking nested ZIP %s: %v\n", file.Name, err)
			} else {
				fmt.Printf("DEBUG: Nested ZIP %s contains %d XCCDF files\n", file.Name, nestedCount)
				totalFiles += nestedCount
			}
		}
	}
	fmt.Printf("DEBUG: Total XCCDF files found: %d\n", totalFiles)

	for _, file := range bundle.File {
		if strings.HasSuffix(strings.ToLower(file.Name), "xccdf.xml") {
			fmt.Printf("DEBUG: Processing XCCDF file: %s\n", file.Name)
			if progressCallback != nil {
				progressCallback(&models.ImportProgress{
					Stage:          "parsing",
					CurrentFile:    file.Name,
					ProcessedFiles: processedFiles,
					TotalFiles:     totalFiles,
					ProcessedSTIGs: 0,
					TotalSTIGs:     0,
					Message:        fmt.Sprintf("Processing: %s", file.Name),
				})
			}

			err := s.processXCCDFFile(tx, file, filename)
			if err != nil {
				fmt.Printf("DEBUG: Error processing XCCDF file %s: %v\n", file.Name, err)
				return fmt.Errorf("failed to process XCCDF file %s: %w", file.Name, err)
			}
			fmt.Printf("DEBUG: Successfully processed XCCDF file: %s\n", file.Name)
			processedFiles++
		} else if strings.HasSuffix(strings.ToLower(file.Name), ".zip") && !file.FileInfo().IsDir() {
			fmt.Printf("DEBUG: Processing nested ZIP file: %s\n", file.Name)
			err := s.processNestedZip(tx, file, filename, progressCallback, &processedFiles, totalFiles)
			if err != nil {
				fmt.Printf("DEBUG: Error processing nested ZIP %s: %v\n", file.Name, err)
				return fmt.Errorf("failed to process nested ZIP %s: %w", file.Name, err)
			}
		}
	}

	fmt.Printf("DEBUG: About to commit transaction. Processed %d files\n", processedFiles)
	// Commit transaction
	if err := tx.Commit(); err != nil {
		fmt.Printf("DEBUG: Failed to commit transaction: %v\n", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	fmt.Printf("DEBUG: Transaction committed successfully\n")

	if progressCallback != nil {
		progressCallback(&models.ImportProgress{
			Stage:   "complete",
			Message: fmt.Sprintf("Successfully imported %d STIGs from bundle", processedFiles),
		})
	}

	return nil
}

// Helper functions
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (s *STIGService) bundleExists(hash string) (bool, error) {
	var count int
	err := s.db.GetConnection().QueryRow("SELECT COUNT(*) FROM bundles WHERE file_hash = ?", hash).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *STIGService) bundleHasSTIGs(filename string) (bool, error) {
	var count int
	err := s.db.GetConnection().QueryRow("SELECT COUNT(*) FROM stigs WHERE filename = ?", filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *STIGService) deleteBundleRecord(hash string) error {
	_, err := s.db.GetConnection().Exec("DELETE FROM bundles WHERE file_hash = ?", hash)
	return err
}

func (s *STIGService) cleanupEmptyBundles(tx *sql.Tx) error {
	// Find bundles that have no associated STIGs and delete them
	_, err := tx.Exec(`
		DELETE FROM bundles
		WHERE file_hash NOT IN (
			SELECT DISTINCT b.file_hash
			FROM bundles b
			INNER JOIN stigs s ON s.filename = b.filename
		)
	`)
	return err
}

// countXCCDFInNestedZip counts XCCDF files in a nested ZIP file
func (s *STIGService) countXCCDFInNestedZip(zipFile *zip.File) (int, error) {
	reader, err := zipFile.Open()
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	// Read the nested ZIP content
	content, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}

	// Create a reader for the nested ZIP
	nestedReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return 0, err
	}

	count := 0
	fmt.Printf("DEBUG: Nested ZIP contains %d files:\n", len(nestedReader.File))
	for i, file := range nestedReader.File {
		fmt.Printf("DEBUG: NestedFile[%d]: %s\n", i, file.Name)
		if strings.HasSuffix(strings.ToLower(file.Name), "xccdf.xml") {
			fmt.Printf("DEBUG: Found XCCDF in nested ZIP: %s\n", file.Name)
			count++
		}
	}

	return count, nil
}

// processNestedZip processes XCCDF files within a nested ZIP file
func (s *STIGService) processNestedZip(tx *sql.Tx, zipFile *zip.File, originalFilename string, progressCallback func(*models.ImportProgress), processedFiles *int, totalFiles int) error {
	reader, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	// Read the nested ZIP content
	content, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Create a reader for the nested ZIP
	nestedReader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return err
	}

	// Process XCCDF files in the nested ZIP
	for _, file := range nestedReader.File {
		if strings.HasSuffix(strings.ToLower(file.Name), "xccdf.xml") {
			fmt.Printf("DEBUG: Processing nested XCCDF file: %s (from %s)\n", file.Name, zipFile.Name)
			if progressCallback != nil {
				progressCallback(&models.ImportProgress{
					Stage:          "parsing",
					CurrentFile:    fmt.Sprintf("%s/%s", zipFile.Name, file.Name),
					ProcessedFiles: *processedFiles,
					TotalFiles:     totalFiles,
					ProcessedSTIGs: 0,
					TotalSTIGs:     0,
					Message:        fmt.Sprintf("Processing: %s", file.Name),
				})
			}

			err := s.processXCCDFFile(tx, file, originalFilename)
			if err != nil {
				fmt.Printf("DEBUG: Error processing nested XCCDF file %s: %v\n", file.Name, err)
				return fmt.Errorf("failed to process nested XCCDF file %s: %w", file.Name, err)
			}
			fmt.Printf("DEBUG: Successfully processed nested XCCDF file: %s\n", file.Name)
			*processedFiles++
		}
	}

	return nil
}

func (s *STIGService) processXCCDFFile(tx *sql.Tx, xccdfFile *zip.File, originalFilename string) error {
	// Extract XCCDF content
	reader, err := xccdfFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open XCCDF file: %w", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read XCCDF content: %w", err)
	}

	// Parse XCCDF
	benchmark, err := parseXCCDF(content)
	if err != nil {
		return fmt.Errorf("failed to parse XCCDF: %w", err)
	}

	// Calculate content hash
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Check if STIG already exists
	exists, err := s.stigExists(tx, benchmark.ID)
	if err != nil {
		return fmt.Errorf("failed to check STIG existence: %w", err)
	}
	if exists {
		// STIG already exists, skip
		return nil
	}

	// Use XML values as primary source, filename as fallback
	version := benchmark.Version
	releaseDate := benchmark.ReleaseDate

	// If version is missing from XML, try to extract from filename
	if version == "" {
		filenameVersion, _ := parseVersionAndReleaseFromFilename(originalFilename)
		version = filenameVersion
	}

	// Extract release from plain-text elements first, then try filename if needed
	if releaseDate == "" {
		// Try to extract release from XCCDF plain-text elements
		releaseDate = extractReleaseFromPlainText(benchmark.PlainTexts)

		// If still not found, try to extract from filename
		if releaseDate == "" {
			_, filenameRelease := parseVersionAndReleaseFromFilename(originalFilename)
			if filenameRelease != "" {
				releaseDate = filenameRelease
			}
		}
	}

	// Insert STIG record
	stigResult, err := tx.Exec(`
		INSERT INTO stigs (name, version, release_date, benchmark_id, title, description, filename, file_hash, rule_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		benchmark.Title, version, releaseDate, benchmark.ID,
		benchmark.Title, benchmark.Description, originalFilename, hash, len(benchmark.Groups),
	)
	if err != nil {
		return fmt.Errorf("failed to insert STIG: %w", err)
	}

	stigID, err := stigResult.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get STIG ID: %w", err)
	}

	// Insert rules
	for _, group := range benchmark.Groups {
		for _, rule := range group.Rules {
			err := s.insertRule(tx, int(stigID), &group, &rule)
			if err != nil {
				return fmt.Errorf("failed to insert rule %s: %w", rule.ID, err)
			}
		}
	}

	return nil
}

func (s *STIGService) stigExists(tx *sql.Tx, benchmarkID string) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM stigs WHERE benchmark_id = ?", benchmarkID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *STIGService) insertRule(tx *sql.Tx, stigID int, group *XCCDFGroup, rule *XCCDFRule) error {
	// Parse CCI refs
	var cciRefs []string
	for _, ident := range rule.Identifiers {
		if ident.System == "http://cyber.mil/cci" {
			cciRefs = append(cciRefs, ident.Text)
		}
	}
	cciJSON := ""
	if len(cciRefs) > 0 {
		if data, err := json.Marshal(cciRefs); err == nil {
			cciJSON = string(data)
		}
	}

	_, err := tx.Exec(`
		INSERT INTO stig_rules (stig_id, rule_id, version_id, title, description, severity, weight, group_id, group_title, check_content, fix_text, cci_refs)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		stigID, rule.ID, rule.Version, rule.Title, rule.Description,
		rule.Severity, rule.Weight, group.ID, group.Title,
		rule.CheckContent, rule.FixText, cciJSON,
	)
	return err
}

// XCCDF parsing structures (simplified)
type XCCDFBenchmark struct {
	XMLName     xml.Name         `xml:"Benchmark"`
	ID          string           `xml:"id,attr"`
	Title       string           `xml:"title"`
	Description string           `xml:"description"`
	Version     string           `xml:"version"`
	ReleaseDate string           `xml:"release-date"`
	PlainTexts  []XCCDFPlainText `xml:"plain-text"`
	Groups      []XCCDFGroup     `xml:"Group"`
}

type XCCDFPlainText struct {
	ID   string `xml:"id,attr"`
	Text string `xml:",chardata"`
}

type XCCDFGroup struct {
	ID          string      `xml:"id,attr"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Rules       []XCCDFRule `xml:"Rule"`
}

type XCCDFRule struct {
	ID           string            `xml:"id,attr"`
	Version      string            `xml:"version,attr"`
	Title        string            `xml:"title"`
	Description  string            `xml:"description"`
	Severity     string            `xml:"severity,attr"`
	Weight       float64           `xml:"weight,attr"`
	CheckContent string            `xml:"check>check-content"`
	FixText      string            `xml:"fixtext"`
	Identifiers  []XCCDFIdentifier `xml:"ident"`
}

type XCCDFIdentifier struct {
	System string `xml:"system,attr"`
	Text   string `xml:",chardata"`
}

// parseXCCDF parses XCCDF content
func parseXCCDF(content []byte) (*XCCDFBenchmark, error) {
	var benchmark XCCDFBenchmark
	err := xml.Unmarshal(content, &benchmark)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XCCDF: %w", err)
	}

	// Debug logging to see what was parsed
	fmt.Printf("DEBUG: Parsed XCCDF - ID: %s, Title: %s, Version: %s, Groups: %d, PlainTexts: %d\n",
		benchmark.ID, benchmark.Title, benchmark.Version, len(benchmark.Groups), len(benchmark.PlainTexts))

	for i, pt := range benchmark.PlainTexts {
		fmt.Printf("DEBUG: PlainText[%d] - ID: %s, Text: %s\n", i, pt.ID, pt.Text)
	}

	return &benchmark, nil
}

// Example test cases for parseVersionAndReleaseFromFilename:
// Input: "U_Active_Directory_Domain_V3R5_STIG" -> Version: "3", Release: "5"
// Input: "U_Windows_10_V2R1_STIG" -> Version: "2", Release: "1"
// Input: "U_Some_STIG_V1_STIG" -> Version: "1", Release: ""
// Input: "U_No_Version_STIG" -> Version: "", Release: ""

// Example test cases for extractReleaseFromPlainText:
// Input: [{ID: "release-info", Text: "Release: 5 Benchmark Date: 13 Sep 2024"}] -> "5"
// Input: [{ID: "release-info", Text: "Release: 12 Benchmark Date: 01 Jan 2025"}] -> "12"
// Input: [{ID: "generator", Text: "3.5"}] -> ""
// Input: [] -> ""
