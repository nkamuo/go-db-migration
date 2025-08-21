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
			// Disable usage on error for clean output
			cmd.SilenceUsage = true

			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				fmt.Printf("❌ Configuration Error\n\n")
				fmt.Printf("Failed to load configuration: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check if conf.json exists in the current directory\n")
				fmt.Printf("   • Verify JSON syntax is valid\n")
				fmt.Printf("   • Use --config flag to specify a different config file\n\n")
				return nil
			}

			// Get connection config
			dbConfig, err := cfg.GetConnectionConfig(connectionName)
			if err != nil {
				fmt.Printf("❌ Connection Configuration Error\n\n")
				fmt.Printf("Failed to get connection config: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check connection name in conf.json\n")
				fmt.Printf("   • Use --connection flag to specify a valid connection\n")
				fmt.Printf("   • Verify default connection is properly configured\n\n")
				return nil
			}

			// Connect to database
			db, err := database.NewConnection(dbConfig)
			if err != nil {
				fmt.Printf("❌ Database Connection Failed\n\n")
				fmt.Printf("Database: %s\n", dbConfig.Database)
				fmt.Printf("Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
				fmt.Printf("User: %s\n\n", dbConfig.Username)
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify database server is running\n")
				fmt.Printf("   • Check connection details in config are correct\n")
				fmt.Printf("   • Ensure user has required permissions\n")
				fmt.Printf("   • Check firewall/network connectivity\n")
				fmt.Printf("   • Verify pg_hba.conf allows your IP address\n\n")
				return nil
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				fmt.Printf("❌ Schema Loading Failed\n\n")
				fmt.Printf("Schema file: %s\n\n", getSchemaFilePath())
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Verify schema file exists and is readable\n")
				fmt.Printf("   • Check JSON format is valid\n")
				fmt.Printf("   • Use --schema flag to specify correct file path\n\n")
				return nil
			}

			// Validate foreign keys
			issues, err := db.ValidateForeignKeys(targetSchema)
			if err != nil {
				fmt.Printf("❌ Foreign Key Validation Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify that all referenced tables exist in the database\n")
				fmt.Printf("   • Check that required columns are present\n")
				fmt.Printf("   • Validate your schema file contains correct foreign key definitions\n")
				fmt.Printf("   • Ensure database connection has proper permissions\n\n")
				fmt.Printf("🔧 Debug Steps:\n")
				fmt.Printf("   1. Run: ./bin/migrator schema info\n")
				fmt.Printf("   2. Check which tables exist in your database\n")
				fmt.Printf("   3. Compare with foreign key references in schema.json\n\n")
				return nil
			}

			// Create report
			report := output.CreateValidationReport(connectionName, issues)

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatValidationReport(report)
			if err != nil {
				fmt.Printf("❌ Output Formatting Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				return nil
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
			// Disable usage on error for clean output
			cmd.SilenceUsage = true

			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				fmt.Printf("❌ Configuration Error\n\n")
				fmt.Printf("Failed to load configuration: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check if conf.json exists in the current directory\n")
				fmt.Printf("   • Verify JSON syntax is valid\n")
				fmt.Printf("   • Use --config flag to specify a different config file\n\n")
				return nil
			}

			// Get connection config
			dbConfig, err := cfg.GetConnectionConfig(connectionName)
			if err != nil {
				fmt.Printf("❌ Connection Configuration Error\n\n")
				fmt.Printf("Failed to get connection config: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check connection name in conf.json\n")
				fmt.Printf("   • Use --connection flag to specify a valid connection\n")
				fmt.Printf("   • Verify default connection is properly configured\n\n")
				return nil
			}

			// Connect to database
			db, err := database.NewConnection(dbConfig)
			if err != nil {
				fmt.Printf("❌ Database Connection Failed\n\n")
				fmt.Printf("Database: %s\n", dbConfig.Database)
				fmt.Printf("Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
				fmt.Printf("User: %s\n\n", dbConfig.Username)
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify database server is running\n")
				fmt.Printf("   • Check connection details in config are correct\n")
				fmt.Printf("   • Ensure user has required permissions\n")
				fmt.Printf("   • Check firewall/network connectivity\n")
				fmt.Printf("   • Verify pg_hba.conf allows your IP address\n\n")
				return nil
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				fmt.Printf("❌ Schema Loading Failed\n\n")
				fmt.Printf("Schema file: %s\n\n", getSchemaFilePath())
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Verify schema file exists and is readable\n")
				fmt.Printf("   • Check JSON format is valid\n")
				fmt.Printf("   • Use --schema flag to specify correct file path\n\n")
				return nil
			}

			// Validate NOT NULL constraints
			issues, err := db.ValidateNotNullConstraints(targetSchema)
			if err != nil {
				fmt.Printf("❌ NOT NULL Validation Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify that target tables exist in the database\n")
				fmt.Printf("   • Check that required columns are present\n")
				fmt.Printf("   • Validate your schema file contains correct column definitions\n")
				fmt.Printf("   • Ensure database connection has proper permissions\n\n")
				fmt.Printf("🔧 Debug Steps:\n")
				fmt.Printf("   1. Run: ./bin/migrator schema info\n")
				fmt.Printf("   2. Check which tables and columns exist in your database\n")
				fmt.Printf("   3. Compare with NOT NULL constraints in schema.json\n\n")
				return nil
			}

			// Create report
			report := output.CreateValidationReport(connectionName, issues)

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatValidationReport(report)
			if err != nil {
				fmt.Printf("❌ Output Formatting Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				return nil
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
			// Disable usage on error for clean output
			cmd.SilenceUsage = true

			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				fmt.Printf("❌ Configuration Error\n\n")
				fmt.Printf("Failed to load configuration: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check if conf.json exists in the current directory\n")
				fmt.Printf("   • Verify JSON syntax is valid\n")
				fmt.Printf("   • Use --config flag to specify a different config file\n\n")
				return nil
			}

			// Get connection config
			dbConfig, err := cfg.GetConnectionConfig(connectionName)
			if err != nil {
				fmt.Printf("❌ Connection Configuration Error\n\n")
				fmt.Printf("Failed to get connection config: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Check connection name in conf.json\n")
				fmt.Printf("   • Use --connection flag to specify a valid connection\n")
				fmt.Printf("   • Verify default connection is properly configured\n\n")
				return nil
			}

			// Connect to database
			db, err := database.NewConnection(dbConfig)
			if err != nil {
				fmt.Printf("❌ Database Connection Failed\n\n")
				fmt.Printf("Database: %s\n", dbConfig.Database)
				fmt.Printf("Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
				fmt.Printf("User: %s\n\n", dbConfig.Username)
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify database server is running\n")
				fmt.Printf("   • Check connection details in config are correct\n")
				fmt.Printf("   • Ensure user has required permissions\n")
				fmt.Printf("   • Check firewall/network connectivity\n")
				fmt.Printf("   • Verify pg_hba.conf allows your IP address\n\n")
				return nil
			}
			defer db.Close()

			// Load target schema
			targetSchema, err := schema.LoadSchema(getSchemaFilePath())
			if err != nil {
				fmt.Printf("❌ Schema Loading Failed\n\n")
				fmt.Printf("Schema file: %s\n\n", getSchemaFilePath())
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Solutions:\n")
				fmt.Printf("   • Verify schema file exists and is readable\n")
				fmt.Printf("   • Check JSON format is valid\n")
				fmt.Printf("   • Use --schema flag to specify correct file path\n\n")
				return nil
			}

			var allIssues []models.ValidationIssue

			// 1. Validate schema structure
			fmt.Println("🔍 Validating schema structure...")
			schemaIssues := schema.ValidateSchema(targetSchema)
			allIssues = append(allIssues, schemaIssues...)

			// 2. Validate foreign keys
			fmt.Println("🔍 Validating foreign key constraints...")
			var fkIssues []models.ValidationIssue
			fkIssues, err = db.ValidateForeignKeys(targetSchema)
			if err != nil {
				fmt.Printf("❌ Foreign Key Validation Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify that all referenced tables exist in the database\n")
				fmt.Printf("   • Check that required columns are present\n")
				fmt.Printf("   • Validate your schema file contains correct foreign key definitions\n\n")
				return nil
			}
			allIssues = append(allIssues, fkIssues...)

			// 3. Validate NOT NULL constraints
			fmt.Println("🔍 Validating NOT NULL constraints...")
			var nullIssues []models.ValidationIssue
			nullIssues, err = db.ValidateNotNullConstraints(targetSchema)
			if err != nil {
				fmt.Printf("❌ NOT NULL Validation Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify that target tables exist in the database\n")
				fmt.Printf("   • Check that required columns are present\n")
				fmt.Printf("   • Validate your schema file contains correct column definitions\n\n")
				return nil
			}
			allIssues = append(allIssues, nullIssues...)

			// Create comprehensive report
			report := output.CreateValidationReport(connectionName, allIssues)

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatValidationReport(report)
			if err != nil {
				fmt.Printf("❌ Output Formatting Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				return nil
			}

			return saveOutput(content, cmd)
		},
	}
}
