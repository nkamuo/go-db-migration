package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/models"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// validateAllCmd represents the validate-all command
var validateAllCmd = &cobra.Command{
	Use:   "validate-all",
	Short: "Run all validation checks",
	Long: `Runs a comprehensive validation that includes:
- Foreign key constraint validation
- NOT NULL constraint validation  
- Schema comparison
- Schema structural validation

This is equivalent to running validate-fk, validate-null, and compare-schema
commands together, providing a complete analysis of migration readiness.`,

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

		// Validate schema structure first
		var allIssues []models.ValidationIssue
		structuralIssues := schema.ValidateSchema(targetSchema)
		allIssues = append(allIssues, structuralIssues...)

		// Validate foreign keys
		fkIssues, err := db.ValidateForeignKeys(targetSchema)
		if err != nil {
			return fmt.Errorf("failed to validate foreign keys: %w", err)
		}
		allIssues = append(allIssues, fkIssues...)

		// Validate NOT NULL constraints
		nullIssues, err := db.ValidateNotNullConstraints(targetSchema)
		if err != nil {
			return fmt.Errorf("failed to validate NOT NULL constraints: %w", err)
		}
		allIssues = append(allIssues, nullIssues...)

		// Create comprehensive report
		report := output.CreateValidationReport(connectionName, allIssues)

		// Also get schema comparison
		currentSchema, err := db.GetCurrentSchema()
		if err != nil {
			return fmt.Errorf("failed to get current schema: %w", err)
		}

		comparison := schema.CompareSchemas(currentSchema, targetSchema)

		// Format and output validation results
		formatter := output.NewFormatter(outputFormat)

		// Output validation report
		validationContent, err := formatter.FormatValidationReport(report)
		if err != nil {
			return fmt.Errorf("failed to format validation report: %w", err)
		}

		// Output schema comparison
		comparisonContent, err := formatter.FormatSchemaComparison(comparison)
		if err != nil {
			return fmt.Errorf("failed to format schema comparison: %w", err)
		}

		// Combine outputs
		combinedContent := fmt.Sprintf("=== VALIDATION REPORT ===\n%s\n\n=== SCHEMA COMPARISON ===\n%s",
			validationContent, comparisonContent)

		return saveOutput(combinedContent, cmd)
	},
}
