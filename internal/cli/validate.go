package cli

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/models"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// newValidateCmd creates the validate command group
func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validation commands for database migration readiness",
		Long: `Commands to validate various aspects of the database to ensure
migration readiness. This includes foreign key constraints, null value
constraints, and comprehensive validation checks.`,
	}

	cmd.AddCommand(newValidateFKCmd())
	cmd.AddCommand(newValidateNullCmd())
	cmd.AddCommand(newValidateAllCmd())

	return cmd
}

// newValidateFKCmd creates the validate fk command
func newValidateFKCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fk",
		Short: "Validate foreign key constraints",
		Long: `Validates foreign key constraints by identifying records that would violate 
these constraints during migration.

This command will:
- Check all foreign key constraints defined in the target schema
- Find orphaned records that reference non-existent parent records
- Provide detailed information including primary keys and identifiers
- Support multiple output formats for easy review and action`,
		Aliases: []string{"foreign-key", "foreign-keys"},

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
				return fmt.Errorf("failed to connect to database '%s': %w\n\nPlease verify:\n- Database server is running\n- Connection details in config are correct\n- User has required permissions", dbConfig.Database, err)
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				return fmt.Errorf("failed to load target schema from '%s': %w\n\nPlease verify:\n- Schema file exists and is readable\n- JSON format is valid", getSchemaFilePath(), err)
			}

			// Validate foreign keys
			issues, err := db.ValidateForeignKeys(targetSchema)
			if err != nil {
				return fmt.Errorf("failed to validate foreign keys: %w\n\nThis error often occurs when:\n- Referenced tables don't exist in the database\n- Required columns are missing\n- Schema file contains invalid foreign key definitions", err)
			} // Create report
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
}

// newValidateNullCmd creates the validate null command
func newValidateNullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "null",
		Short: "Validate NOT NULL constraints",
		Long: `Validates NOT NULL constraints by identifying records with null values 
in columns that will be made NOT NULL during migration.

This command will:
- Check all columns marked as NOT NULL in the target schema
- Find records with null values in these columns
- Provide detailed information including primary keys and identifiers
- Support multiple output formats for easy review and action`,
		Aliases: []string{"not-null", "nulls"},

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
				return fmt.Errorf("failed to connect to database '%s': %w\n\nPlease verify:\n- Database server is running\n- Connection details in config are correct\n- User has required permissions", dbConfig.Database, err)
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				return fmt.Errorf("failed to load target schema from '%s': %w\n\nPlease verify:\n- Schema file exists and is readable\n- JSON format is valid", getSchemaFilePath(), err)
			}

			// Validate NOT NULL constraints
			issues, err := db.ValidateNotNullConstraints(targetSchema)
			if err != nil {
				return fmt.Errorf("failed to validate NOT NULL constraints: %w\n\nThis error often occurs when:\n- Target tables don't exist in the database\n- Required columns are missing\n- Schema file contains invalid column definitions", err)
			} // Create report
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
}

// newValidateAllCmd creates the validate all command
func newValidateAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all",
		Short: "Run all validation checks",
		Long: `Runs all available validation checks including foreign key constraints,
NOT NULL constraints, and schema validation.

This is a comprehensive check that combines:
- Foreign key constraint validation
- NOT NULL constraint validation
- Schema structure validation
- Data integrity checks`,

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
				return fmt.Errorf("failed to connect to database '%s': %w\n\nPlease verify:\n- Database server is running\n- Connection details in config are correct\n- User has required permissions", dbConfig.Database, err)
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				return fmt.Errorf("failed to load target schema from '%s': %w\n\nPlease verify:\n- Schema file exists and is readable\n- JSON format is valid", getSchemaFilePath(), err)
			}

			var allIssues []models.ValidationIssue

			// 1. Validate schema structure
			fmt.Println("🔍 Validating schema structure...")
			schemaIssues := schema.ValidateSchema(targetSchema)
			allIssues = append(allIssues, schemaIssues...)

			// 2. Validate foreign keys
			fmt.Println("🔍 Validating foreign key constraints...")
			fkIssues, err := db.ValidateForeignKeys(targetSchema)
			if err != nil {
				return fmt.Errorf("failed to validate foreign keys: %w\n\nThis error often occurs when:\n- Referenced tables don't exist in the database\n- Required columns are missing\n- Schema file contains invalid foreign key definitions", err)
			}
			allIssues = append(allIssues, fkIssues...)

			// 3. Validate NOT NULL constraints
			fmt.Println("🔍 Validating NOT NULL constraints...")
			nullIssues, err := db.ValidateNotNullConstraints(targetSchema)
			if err != nil {
				return fmt.Errorf("failed to validate NOT NULL constraints: %w\n\nThis error often occurs when:\n- Target tables don't exist in the database\n- Required columns are missing\n- Schema file contains invalid column definitions", err)
			}
			allIssues = append(allIssues, nullIssues...)

			// Create comprehensive report
			report := output.CreateValidationReport(connectionName, allIssues)

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatValidationReport(report)
			if err != nil {
				return fmt.Errorf("failed to format output: %w", err)
			}

			return saveOutput(content, cmd)
		},
	}
}
