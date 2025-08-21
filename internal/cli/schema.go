package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
)

// newSchemaCmd creates the schema command group
func newSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Schema analysis and comparison commands",
		Long: `Commands for analyzing and comparing database schemas.
This includes comparing current database schema with target schema,
validating schema files, and generating schema reports.`,
	}

	cmd.AddCommand(newSchemaCompareCmd())
	cmd.AddCommand(newSchemaValidateCmd())
	cmd.AddCommand(newSchemaInfoCmd())

	return cmd
}

// newSchemaCompareCmd creates the schema compare command
func newSchemaCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare",
		Short: "Compare current database schema with target schema",
		Long: `Compares the current database schema with the target schema file
to identify structural differences.

This command will:
- Compare table structures between current and target schemas
- Identify missing, extra, or modified tables
- Compare column definitions and constraints
- Highlight foreign key differences
- Support multiple output formats for detailed analysis`,
		Aliases: []string{"diff", "compare-schema"},

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
}

// newSchemaValidateCmd creates the schema validate command
func newSchemaValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the target schema file",
		Long: `Validates the target schema file for structural consistency and correctness.
This includes checking for:
- Duplicate table or column names
- Missing required fields
- Invalid foreign key references
- Schema format validation`,
		Aliases: []string{"check"},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				return fmt.Errorf("failed to load target schema: %w", err)
			}

			// Validate schema structure
			issues := schema.ValidateSchema(targetSchema)

			if len(issues) == 0 {
				fmt.Println("✅ Schema file is valid!")
				fmt.Printf("Found %d tables in schema\n", len(targetSchema))
				return nil
			}

			fmt.Printf("❌ Found %d validation issues in schema:\n", len(issues))
			for _, issue := range issues {
				fmt.Printf("  [%s] %s: %s\n", issue.Severity, issue.Type, issue.Message)
				if issue.Table != "" {
					fmt.Printf("    Table: %s\n", issue.Table)
				}
				if issue.Column != "" {
					fmt.Printf("    Column: %s\n", issue.Column)
				}
			}

			return fmt.Errorf("schema validation failed")
		},
	}
}

// newSchemaInfoCmd creates the schema info command
func newSchemaInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display information about the target schema",
		Long: `Displays detailed information about the target schema including:
- Number of tables and columns
- Foreign key relationships
- Data types summary
- Schema statistics`,
		Aliases: []string{"stats", "summary"},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				return fmt.Errorf("failed to load target schema: %w", err)
			}

			// Create schema info
			schemaInfo := output.CreateSchemaInfo(getSchemaFilePath(), targetSchema)

			// Format output
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatSchemaInfo(schemaInfo)
			if err != nil {
				return fmt.Errorf("failed to format schema info: %w", err)
			}

			// Output or save to file
			if outputFile != "" {
				return output.WriteToFile(content, outputFile)
			}

			fmt.Print(content)
			return nil
		},
	}
}
