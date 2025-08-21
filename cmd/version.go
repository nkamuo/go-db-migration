package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "1.0.0"
var BuildDate = ""
var GitCommit = ""

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display the version, build date, and git commit information for this tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("stsx-db-migration version %s\n", Version)
		if BuildDate != "" {
			fmt.Printf("Build date: %s\n", BuildDate)
		}
		if GitCommit != "" {
			fmt.Printf("Git commit: %s\n", GitCommit)
		}
	},
}
