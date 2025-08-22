package cli

import (
	"fmt"
	"os"

	"github.com/nkamuo/go-db-migration/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile        string
	connectionName string
	schemaFile     string
	outputFormat   string
	outputFile     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "migrator",
	Short: "Database Migration Validator - A robust tool for validating database schemas",
	Long: `A robust database validation tool designed to identify potential issues 
that could prevent successful database migrations.

This tool can:
- Validate foreign key constraints
- Check for null values in columns that will be made NOT NULL
- Compare current database schema with target schema
- Generate detailed reports in multiple formats (table, JSON, YAML, CSV)

Examples:
  # Validate foreign keys using default connection
  migrator validate fk

  # Check null constraints with specific connection
  migrator validate null --connection "JAMES Database"

  # Compare schemas and output to JSON
  migrator schema compare --format json --output report.json

  # Test database connection
  migrator connection test --connection "JAMES Database"`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip validation for help, version, and schema info commands
		if cmd.Name() == "help" || cmd.Name() == "version" ||
			(cmd.Parent() != nil && cmd.Parent().Name() == "schema" && cmd.Name() == "info") ||
			(cmd.Parent() != nil && cmd.Parent().Name() == "schema" && cmd.Name() == "validate") {
			return nil
		}

		// Load and validate configuration
		cfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Store config in context for subcommands
		cmd.SetContext(cmd.Context())
		cmd.Annotations = map[string]string{"config": fmt.Sprintf("%+v", cfg)}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./conf.json)")
	rootCmd.PersistentFlags().StringVarP(&connectionName, "connection", "c", "", "database connection name from config")
	rootCmd.PersistentFlags().StringVarP(&schemaFile, "schema", "s", "", "target schema file (default is ./schema.json)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "table", "output format (table, json, yaml, csv)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file (default is stdout)")

	// Add command groups
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newSchemaCmd())
	rootCmd.AddCommand(newConnectionCmd())
	rootCmd.AddCommand(newFixCmd())
	rootCmd.AddCommand(newVersionCmd())
}

// Helper function to get config from command context
func getConfigFromCmd(cmd *cobra.Command) (*config.Config, error) {
	return config.LoadConfig(cfgFile)
}

// Helper function to get schema file path
func getSchemaFilePath() string {
	if schemaFile != "" {
		return schemaFile
	}

	// Try current directory first
	if _, err := os.Stat("./schema.json"); err == nil {
		return "./schema.json"
	}

	// Fall back to default path
	return config.GetDefaultSchemaPath()
}

// Helper function to save output
func saveOutput(content string, cmd *cobra.Command) error {
	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(content), 0644)
	}

	// Print to stdout
	fmt.Print(content)
	return nil
}
