package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// validateNullCmd represents the validate-null command
var validateNullCmd = &cobra.Command{
	Use:   "validate-null",
	Short: "Validate NOT NULL constraints",
	Long: `Validates NOT NULL constraints by identifying records with null values 
in columns that will be made NOT NULL during migration.

This command will:
- Check all columns marked as NOT NULL in the target schema
- Find records with null values in these columns
- Provide detailed information including primary keys and identifiers
- Support multiple output formats for easy review and action`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration
		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}

		// Get connection config
		dbConfig, err := cfg.GetConnectionConfig(connectionName)
		if err != nil {
			return err
		}

		// Connect to database
		db, err := database.NewConnection(dbConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer db.Close()

		// Load target schema
		targetSchema, err := schema.LoadSchema(getSchemaFilePath())
		if err != nil {
			return fmt.Errorf("failed to load target schema: %w", err)
		}

		// Validate NOT NULL constraints
		issues, err := db.ValidateNotNullConstraints(targetSchema)
		if err != nil {
			return fmt.Errorf("failed to validate NOT NULL constraints: %w", err)
		}

		// Create report
		report := output.CreateValidationReport(connectionName, issues)

		// Format and output results
		formatter := output.NewFormatter(outputFormat)
		content, err := formatter.FormatValidationReport(report)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		return saveOutput(content, cmd)
	},
}
