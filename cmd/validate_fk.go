package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// validateFKCmd represents the validate-fk command
var validateFKCmd = &cobra.Command{
	Use:   "validate-fk",
	Short: "Validate foreign key constraints",
	Long: `Validates foreign key constraints by identifying records that would violate 
these constraints during migration.

This command will:
- Check all foreign key constraints defined in the target schema
- Find orphaned records that reference non-existent parent records
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

		// Validate foreign keys
		issues, err := db.ValidateForeignKeys(targetSchema)
		if err != nil {
			return fmt.Errorf("failed to validate foreign keys: %w", err)
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
