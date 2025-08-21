package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// These will be set during build time
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long: `Displays version information including the application version,
Git commit hash, build date, and Go runtime information.`,

		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Database Migrator\n")
			fmt.Printf("================\n\n")
			fmt.Printf("Version:    %s\n", Version)
			fmt.Printf("Git Commit: %s\n", GitCommit)
			fmt.Printf("Build Date: %s\n", BuildDate)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
