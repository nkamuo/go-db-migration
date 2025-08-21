package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// compareSchemaCmd represents the compare-schema command
var compareSchemaCmd = &cobra.Command{
	Use:   "compare-schema",
	Short: "Compare current database schema with target schema",
	Long: `Compares the current database schema with the target schema file
to identify structural differences.

This command will:
- Compare table structures between current and target schemas
- Identify missing, extra, or modified tables
- Compare column definitions and constraints
- Highlight foreign key differences
- Support multiple output formats for detailed analysis`,

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

		// Get current schema
		currentSchema, err := db.GetCurrentSchema()
		if err != nil {
			return fmt.Errorf("failed to get current schema: %w", err)
		}

		// Load target schema
		targetSchema, err := schema.LoadSchema(getSchemaFilePath())
		if err != nil {
			return fmt.Errorf("failed to load target schema: %w", err)
		}

		// Compare schemas
		comparison := schema.CompareSchemas(currentSchema, targetSchema)

		// Format and output results
		formatter := output.NewFormatter(outputFormat)
		content, err := formatter.FormatSchemaComparison(comparison)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}

		return saveOutput(content, cmd)
	},
}
