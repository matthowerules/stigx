package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stigx",
	Short: "STIG Viewer X - A cross-platform STIG/CKL viewer and editor",
	Long: `STIG Viewer X is a modern, cross-platform tool for viewing and editing 
DISA STIGs (Security Technical Implementation Guides) and CKL/CKLb checklists.

Features:
- View and edit CKL (XML) and CKLb (JSON) checklists
- Convert between CKL and CKLb formats
- Import STIG bundles and XCCDF content
- Cross-platform support (Windows, macOS, Linux)
- Modern GUI with Wails and CLI interface`,
	Version: "0.1.0-dev",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}