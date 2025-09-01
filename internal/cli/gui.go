package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long: `Launch the STIG Viewer X graphical user interface built with Wails.
The GUI provides a modern, user-friendly interface for viewing and editing
STIG checklists with features like drag-and-drop file handling, search,
filtering, and visual diff tools.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Launching STIG Viewer X GUI...")
		// TODO: Implement Wails GUI launch
		fmt.Println("GUI functionality not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
	guiCmd.Flags().StringP("file", "f", "", "Open a specific checklist file on startup")
	guiCmd.Flags().BoolP("dev", "d", false, "Launch in development mode with devtools")
}