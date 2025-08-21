package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// validateSchemaCmd represents the validate-schema command
var validateSchemaCmd = &cobra.Command{
	Use:   "validate-schema",
	Short: "Validate the target schema file",
	Long: `Validates the target schema file for structural consistency and correctness.
This includes checking for:
- Duplicate table or column names
- Missing required fields
- Invalid foreign key references
- Schema format validation`,

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

func init() {
	rootCmd.AddCommand(validateSchemaCmd)
}
