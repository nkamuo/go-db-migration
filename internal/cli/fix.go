package cli

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
)

// Fix command options
var (
	dryRun         bool
	fixAction      string
	defaultValue   string
	confirmChanges bool
)

// newFixCmd creates the fix command group
func newFixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix database issues found during validation",
		Long: `Commands to automatically fix common database issues that would prevent
successful migration. This includes cleaning up foreign key violations and
null value issues.

⚠️  WARNING: These commands modify your database. Always run with --dry-run first
and backup your data before running actual fixes.`,
	}

	cmd.AddCommand(newFixFKCmd())
	cmd.AddCommand(newFixNullCmd())

	// Add persistent flags
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making actual changes")
	cmd.PersistentFlags().BoolVar(&confirmChanges, "confirm", false, "Confirm that you want to make actual changes (required for non-dry-run)")

	return cmd
}

// newFixFKCmd creates the fix foreign key command
func newFixFKCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fk",
		Short: "Fix foreign key constraint violations",
		Long: `Fixes foreign key constraint violations by either removing orphaned records
or setting foreign key columns to NULL.

Available actions:
  - remove: Delete records that violate foreign key constraints
  - set-null: Set foreign key columns to NULL for violating records

Examples:
  migrator fix fk --action remove --dry-run
  migrator fix fk --action set-null --confirm`,
		Aliases: []string{"foreign-key", "foreign-keys"},

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if fixAction == "" {
				return fmt.Errorf("--action is required (remove|set-null)")
			}

			if fixAction != "remove" && fixAction != "set-null" {
				return fmt.Errorf("invalid action: %s (must be 'remove' or 'set-null')", fixAction)
			}

			// Handle dry-run defaults: if neither --dry-run nor --confirm is explicitly set,
			// default to dry-run for safety
			if !cmd.Flags().Changed("dry-run") && !cmd.Flags().Changed("confirm") {
				dryRun = true
			}

			// If --confirm is set, disable dry-run (unless --dry-run is explicitly set)
			if confirmChanges && !cmd.Flags().Changed("dry-run") {
				dryRun = false
			}

			if !dryRun && !confirmChanges {
				return fmt.Errorf("must use --confirm flag when not in dry-run mode")
			}

			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Get connection config
			dbConfig, err := cfg.GetConnectionConfig(connectionName)
			if err != nil {
				return fmt.Errorf("failed to get connection config: %w", err)
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

			// Get validation config
			validationConfig := getValidationConfigFromFlags()

			fmt.Printf("🔧 Foreign Key Constraint Fix\n")
			fmt.Printf("   Database: %s\n", dbConfig.Database)
			fmt.Printf("   Action: %s\n", fixAction)
			fmt.Printf("   Dry Run: %v\n", dryRun)
			fmt.Printf("\n")

			if dryRun {
				fmt.Printf("🔍 Analyzing foreign key violations (dry-run mode)...\n")
			} else {
				fmt.Printf("⚠️  MAKING ACTUAL CHANGES TO DATABASE!\n")
				fmt.Printf("🔧 Fixing foreign key violations...\n")
			}

			// Fix foreign key issues
			results, err := db.FixForeignKeyViolations(targetSchema, fixAction, dryRun, &validationConfig)
			if err != nil {
				return fmt.Errorf("failed to fix foreign key violations: %w", err)
			}

			// Display results
			fmt.Printf("\n📊 Fix Results:\n")
			for tableName, result := range results {
				fmt.Printf("  Table: %s\n", tableName)
				fmt.Printf("    Issues found: %d\n", result.IssuesFound)
				fmt.Printf("    Records affected: %d\n", result.RecordsAffected)
				if !dryRun {
					fmt.Printf("    Changes applied: %v\n", result.Success)
				}
			}

			if dryRun {
				fmt.Printf("\n💡 To apply these changes, run with --confirm flag and without --dry-run\n")
			} else {
				fmt.Printf("\n✅ Fix operation completed!\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&fixAction, "action", "", "Fix action: remove|set-null")

	return cmd
}

// newFixNullCmd creates the fix null values command
func newFixNullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "null",
		Short: "Fix NULL value constraint violations",
		Long: `Fixes NULL value issues in columns that will be set to NOT NULL during migration.

Available actions:
  - remove: Delete records with NULL values in target columns
  - set-default: Set NULL values to specified default value

Examples:
  migrator fix null --action remove --dry-run
  migrator fix null --action set-default --default-value "unknown" --confirm`,
		Aliases: []string{"not-null", "nulls"},

		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if fixAction == "" {
				return fmt.Errorf("--action is required (remove|set-default)")
			}

			if fixAction != "remove" && fixAction != "set-default" {
				return fmt.Errorf("invalid action: %s (must be 'remove' or 'set-default')", fixAction)
			}

			if fixAction == "set-default" && defaultValue == "" {
				return fmt.Errorf("--default-value is required when using 'set-default' action")
			}

			// Handle dry-run defaults: if neither --dry-run nor --confirm is explicitly set,
			// default to dry-run for safety
			if !cmd.Flags().Changed("dry-run") && !cmd.Flags().Changed("confirm") {
				dryRun = true
			}

			// If --confirm is set, disable dry-run (unless --dry-run is explicitly set)
			if confirmChanges && !cmd.Flags().Changed("dry-run") {
				dryRun = false
			}

			if !dryRun && !confirmChanges {
				return fmt.Errorf("must use --confirm flag when not in dry-run mode")
			}

			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Get connection config
			dbConfig, err := cfg.GetConnectionConfig(connectionName)
			if err != nil {
				return fmt.Errorf("failed to get connection config: %w", err)
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

			// Get validation config
			validationConfig := getValidationConfigFromFlags()

			fmt.Printf("🔧 NULL Value Fix\n")
			fmt.Printf("   Database: %s\n", dbConfig.Database)
			fmt.Printf("   Action: %s\n", fixAction)
			if fixAction == "set-default" {
				fmt.Printf("   Default Value: %s\n", defaultValue)
			}
			fmt.Printf("   Dry Run: %v\n", dryRun)
			fmt.Printf("\n")

			if dryRun {
				fmt.Printf("🔍 Analyzing NULL value violations (dry-run mode)...\n")
			} else {
				fmt.Printf("⚠️  MAKING ACTUAL CHANGES TO DATABASE!\n")
				fmt.Printf("🔧 Fixing NULL value violations...\n")
			}

			// Fix null value issues
			results, err := db.FixNullValueViolations(targetSchema, fixAction, defaultValue, dryRun, &validationConfig)
			if err != nil {
				return fmt.Errorf("failed to fix null value violations: %w", err)
			}

			// Display results
			fmt.Printf("\n📊 Fix Results:\n")
			for tableName, result := range results {
				fmt.Printf("  Table: %s\n", tableName)
				fmt.Printf("    Issues found: %d\n", result.IssuesFound)
				fmt.Printf("    Records affected: %d\n", result.RecordsAffected)
				if !dryRun {
					fmt.Printf("    Changes applied: %v\n", result.Success)
				}
			}

			if dryRun {
				fmt.Printf("\n💡 To apply these changes, run with --confirm flag and without --dry-run\n")
			} else {
				fmt.Printf("\n✅ Fix operation completed!\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&fixAction, "action", "", "Fix action: remove|set-default")
	cmd.Flags().StringVar(&defaultValue, "default-value", "", "Default value to use when action is set-default")

	return cmd
}
