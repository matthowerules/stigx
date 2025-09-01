package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/matthowerules/stigx/internal/converters"
	"github.com/matthowerules/stigx/internal/database"
	"github.com/matthowerules/stigx/internal/models"
	"github.com/matthowerules/stigx/internal/parsers"
	"github.com/matthowerules/stigx/internal/services"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct contains the application context and services
type App struct {
	ctx               context.Context
	detector          *parsers.FileDetector
	converter         *converters.Converter
	db                *database.DB
	stigService       *services.STIGService
	checklistService  *services.ChecklistService
	securityValidator *models.SecurityValidator
	errorHandler      *models.ErrorHandler
}

// NewApp creates a new App application struct
func NewApp() (*App, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, models.FileError{
			Operation: "get config directory",
			FilePath:  "user config",
			Cause:     err,
		}
	}

	// Create app data directory
	appDataDir := filepath.Join(configDir, "stigx")
	if err := os.MkdirAll(appDataDir, 0755); err != nil {
		return nil, models.FileError{
			Operation: "create",
			FilePath:  appDataDir,
			Cause:     err,
		}
	}

	// Initialize database
	dbPath := filepath.Join(appDataDir, "stigx.db")
	db, err := database.New(dbPath)
	if err != nil {
		return nil, models.DatabaseError{
			Operation: "initialize",
			Table:     "all",
			Cause:     err,
		}
	}

	// Initialize services
	stigService := services.NewSTIGService(db)
	checklistService := services.NewChecklistService(db, stigService)

	return &App{
		detector:          parsers.NewFileDetector(),
		converter:         converters.NewConverter(),
		db:                db,
		stigService:       stigService,
		checklistService:  checklistService,
		securityValidator: models.NewSecurityValidator(),
		errorHandler:      models.NewErrorHandler(),
	}, nil
}

// Startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// FileInfo represents file information for the frontend
type FileInfo struct {
	Name            string                 `json:"name"`
	Path            string                 `json:"path"`
	Size            int64                  `json:"size"`
	Format          string                 `json:"format"`
	IsValid         bool                   `json:"isValid"`
	ValidationError string                 `json:"validationError,omitempty"`
	Statistics      map[string]interface{} `json:"statistics,omitempty"`
}

// ConversionResult represents the result of a file conversion
type ConversionResult struct {
	Success      bool   `json:"success"`
	OutputPath   string `json:"outputPath,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	InputSize    int64  `json:"inputSize"`
	OutputSize   int64  `json:"outputSize"`
}

// OpenFileDialog opens a file selection dialog
func (a *App) OpenFileDialog() (string, error) {
	filters := []runtime.FileFilter{
		{
			DisplayName: "STIG Files",
			Pattern:     "*.ckl;*.cklb;*.json;*.xml",
		},
		{
			DisplayName: "CKL Files (XML)",
			Pattern:     "*.ckl;*.xml",
		},
		{
			DisplayName: "CKLb Files (JSON)",
			Pattern:     "*.cklb;*.json",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*",
		},
	}

	options := runtime.OpenDialogOptions{
		Title:   "Open CKL File",
		Filters: filters,
	}

	filePath, err := runtime.OpenFileDialog(a.ctx, options)
	if err != nil {
		return "", models.FileError{
			Operation: "open dialog",
			FilePath:  "user selection",
			Cause:     err,
		}
	}

	// Validate the selected file path for security
	if filePath != "" {
		if err := a.securityValidator.ValidateFilePath(filePath); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// ImportStigBundleDialog opens a file selection dialog for STIG bundles
func (a *App) ImportStigBundleDialog() (string, error) {
	filters := []runtime.FileFilter{
		{
			DisplayName: "STIG Bundle Files",
			Pattern:     "*.zip",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*",
		},
	}

	options := runtime.OpenDialogOptions{
		Title:   "Import STIG Bundle",
		Filters: filters,
	}

	return runtime.OpenFileDialog(a.ctx, options)
}

// ImportStigBundle imports a STIG bundle zip file
func (a *App) ImportStigBundle(bundlePath string) error {
	return a.stigService.ImportBundle(bundlePath, func(progress *models.ImportProgress) {
		// Send progress updates to frontend via runtime events
		runtime.EventsEmit(a.ctx, "import-progress", progress)
	})
}

// GetImportedStigs returns all imported STIGs
func (a *App) GetImportedStigs() ([]*models.ImportedSTIGSummary, error) {
	return a.stigService.GetAllSTIGs()
}

// DeleteImportedStigs deletes multiple imported STIGs by their IDs
func (a *App) DeleteImportedStigs(stigIds []int) error {
	return a.stigService.DeleteSTIGs(stigIds)
}

// CreateChecklist creates a new checklist from selected STIGs
func (a *App) CreateChecklist(req *services.CreateChecklistRequest) (*services.ChecklistCreationResult, error) {
	return a.checklistService.CreateChecklist(req)
}

// CreateChecklistFromStigs creates a full checklist data structure from selected STIGs
func (a *App) CreateChecklistFromStigs(req *services.CreateChecklistRequest) (*services.ChecklistWithItems, error) {
	return a.checklistService.CreateChecklistFromStigs(req)
}

// LoadChecklistFromFile loads a checklist from a CKL or CKLb file
func (a *App) LoadChecklistFromFile(filePath string) (*services.ChecklistWithItems, error) {
	return a.checklistService.LoadChecklistFromFile(filePath)
}

// SaveChecklistToFile saves a checklist to a file
func (a *App) SaveChecklistToFile(checklistData *services.ChecklistWithItems, format string, outputPath string) error {
	return a.checklistService.GenerateChecklistFile(checklistData, format, outputPath)
}

// SaveFileDialog opens a file save dialog
func (a *App) SaveFileDialog(defaultName string) (string, error) {
	// Determine preferred format based on default filename extension
	isCKLb := strings.HasSuffix(strings.ToLower(defaultName), ".cklb")

	var filters []runtime.FileFilter
	if isCKLb {
		// Put CKLb first for checklist creation (defaults to CKLb)
		filters = []runtime.FileFilter{
			{
				DisplayName: "CKLb Files (JSON)",
				Pattern:     "*.cklb",
			},
			{
				DisplayName: "CKL Files (XML)",
				Pattern:     "*.ckl",
			},
		}
	} else {
		// Put CKL first for other operations
		filters = []runtime.FileFilter{
			{
				DisplayName: "CKL Files (XML)",
				Pattern:     "*.ckl",
			},
			{
				DisplayName: "CKLb Files (JSON)",
				Pattern:     "*.cklb",
			},
		}
	}

	// Sanitize the default filename for security
	sanitizedName := a.securityValidator.SanitizeFileName(defaultName)

	options := runtime.SaveDialogOptions{
		Title:           "Save Converted File",
		DefaultFilename: sanitizedName,
		Filters:         filters,
	}

	filePath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", models.FileError{
			Operation: "save dialog",
			FilePath:  "user selection",
			Cause:     err,
		}
	}

	// Validate the selected file path for security
	if filePath != "" {
		if err := a.securityValidator.ValidateFilePath(filePath); err != nil {
			return "", err
		}
	}

	return filePath, nil
}

// AnalyzeFile analyzes a file and returns information about it
func (a *App) AnalyzeFile(filePath string) (*FileInfo, error) {
	info := &FileInfo{
		Name: filepath.Base(filePath),
		Path: filePath,
	}

	// Get file info from detector
	fileInfo, err := a.detector.GetFileInfo(filePath)
	if err != nil {
		info.IsValid = false
		info.ValidationError = err.Error()
		return info, nil // Don't wrap error here as it's already structured
	}

	// Extract information
	if size, ok := fileInfo["file_size"].(int64); ok {
		info.Size = size
	}

	if format, ok := fileInfo["detected_format"].(string); ok {
		info.Format = format
		info.IsValid = true
	} else {
		info.Format = "Unknown"
		info.IsValid = false
	}

	if parseError, ok := fileInfo["parse_error"].(string); ok {
		info.IsValid = false
		info.ValidationError = parseError
	}

	if stats, ok := fileInfo["statistics"].(map[string]interface{}); ok {
		info.Statistics = stats
	}

	return info, nil
}

// ValidateFile validates a STIG file and returns detailed information
func (a *App) ValidateFile(filePath string) (*FileInfo, error) {
	// First analyze the file
	info, err := a.AnalyzeFile(filePath)
	if err != nil {
		return nil, err
	}

	// If basic analysis failed, return early
	if !info.IsValid {
		return info, nil
	}

	// Perform detailed validation based on format
	fileType, err := a.detector.DetectFile(filePath)
	if err != nil {
		info.IsValid = false
		info.ValidationError = err.Error()
		return info, nil
	}

	switch fileType {
	case parsers.FileTypeCKL:
		parser := parsers.NewCKLParser()
		_, err := parser.ParseFile(filePath)
		if err != nil {
			info.IsValid = false
			info.ValidationError = err.Error()
		}
	case parsers.FileTypeCKLb:
		parser := parsers.NewCKLbParser()
		_, err := parser.ParseFile(filePath)
		if err != nil {
			info.IsValid = false
			info.ValidationError = err.Error()
		}
	default:
		info.IsValid = false
		info.ValidationError = "Unsupported file format"
	}

	return info, nil
}

// ConvertFile converts a file from one format to another
func (a *App) ConvertFile(inputPath, outputPath string) (*ConversionResult, error) {
	result := &ConversionResult{}

	// Get input file size
	inputInfo, err := a.detector.GetFileInfo(inputPath)
	if err != nil {
		result.Success = false
		result.ErrorMessage = fmt.Sprintf("Failed to read input file: %v", err)
		return result, nil // Don't wrap error here as it's already structured
	}

	if size, ok := inputInfo["file_size"].(int64); ok {
		result.InputSize = size
	}

	// Perform conversion
	err = a.converter.ConvertFile(inputPath, outputPath)
	if err != nil {
		result.Success = false
		result.ErrorMessage = err.Error()
		return result, nil // Don't wrap error here as it's already structured
	}

	// Get output file size
	outputInfo, err := a.detector.GetFileInfo(outputPath)
	if err != nil {
		result.Success = true // Conversion succeeded even if we can't get output size
		result.OutputPath = outputPath
		return result, nil
	}

	if size, ok := outputInfo["file_size"].(int64); ok {
		result.OutputSize = size
	}

	result.Success = true
	result.OutputPath = outputPath
	return result, nil
}

// GetAppInfo returns information about the application
func (a *App) GetAppInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "STIG Viewer X",
		"version":     "0.1.0-dev",
		"description": "Modern cross-platform STIG/CKL viewer and editor",
		"features": []string{
			"CKL/CKLb file validation",
			"Format conversion (CKL ⇄ CKLb)",
			"STIG rule viewing and editing",
			"Cross-platform support",
		},
	}
}

func main() {
	// Create an instance of the app structure
	app, err := NewApp()
	if err != nil {
		println("Error creating app:", err.Error())
		return
	}

	// Ensure database is closed on exit
	defer func() {
		if app.db != nil {
			app.db.Close()
		}
	}()

	// Create application with options
	err = wails.Run(&options.App{
		Title:     "STIG Viewer X",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		OnBeforeClose: func(ctx context.Context) bool {
			return false
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
