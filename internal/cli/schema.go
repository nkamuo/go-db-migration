package cli

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/nkamuo/go-db-migration/internal/output"
	"github.com/nkamuo/go-db-migration/internal/schema"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(newSchemaExportCmd())
	cmd.AddCommand(newSchemaSnapshotCmd())

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

// newSchemaExportCmd creates the schema export command
func newSchemaExportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export database schema to file",
		Long: `Export the current database schema to a JSON file or other formats.
This command connects to the database and extracts the complete schema including:
- Table structures with vendor-specific data types and proper sizing (e.g., 'character varying(50)')
- Column definitions with constraints, nullability, and default values
- Foreign key relationships with referential integrity rules
- Detailed metadata for comprehensive analysis

The exported schema can be used as input for validation and comparison commands.
For a simplified snapshot format, use the 'schema snapshot' command instead.`,
		Aliases: []string{"dump", "extract"},

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

			// Export current schema
			fmt.Printf("🔄 Exporting schema from database '%s'...\n", dbConfig.Database)
			currentSchema, err := db.GetCurrentSchema()
			if err != nil {
				fmt.Printf("❌ Schema Export Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify user has permission to read schema information\n")
				fmt.Printf("   • Check if database contains tables in 'public' schema\n")
				fmt.Printf("   • Ensure database connection is stable\n\n")
				return nil
			}

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatSchema(currentSchema)
			if err != nil {
				fmt.Printf("❌ Output Formatting Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				return nil
			}

			// Save or print output
			err = saveOutput(content, cmd)
			if err != nil {
				fmt.Printf("❌ Failed to save output: %v\n", err)
				return nil
			}

			fmt.Printf("✅ Schema exported successfully\n")
			if outputPath := cmd.Flag("output").Value.String(); outputPath != "" {
				fmt.Printf("📁 Saved to: %s\n", outputPath)
			}

			return nil
		},
	}
}

// newSchemaSnapshotCmd creates the schema snapshot command
func newSchemaSnapshotCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "snapshot",
		Short: "Create a simplified schema snapshot",
		Long: `Create a simplified schema snapshot showing only table names and column types.
This command generates a compact schema representation that focuses on the basic structure
without detailed metadata. Useful for quick schema comparisons and version tracking.

The snapshot format includes:
- Table names and column names
- Column data types with proper vendor-specific formatting (e.g., 'character varying(50)')
- Minimal structure for easy comparison and version control`,
		Aliases: []string{"snap", "simple"},

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

			// Export current schema
			fmt.Printf("📸 Creating schema snapshot from database '%s'...\n", dbConfig.Database)
			currentSchema, err := db.GetCurrentSchema()
			if err != nil {
				fmt.Printf("❌ Schema Snapshot Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				fmt.Printf("💡 Common Solutions:\n")
				fmt.Printf("   • Verify user has permission to read schema information\n")
				fmt.Printf("   • Check if database contains tables in 'public' schema\n")
				fmt.Printf("   • Ensure database connection is stable\n\n")
				return nil
			}

			// Format and output results
			formatter := output.NewFormatter(outputFormat)
			content, err := formatter.FormatSchemaSnapshot(currentSchema)
			if err != nil {
				fmt.Printf("❌ Output Formatting Failed\n\n")
				fmt.Printf("Error: %v\n\n", err)
				return nil
			}

			// Save or print output
			err = saveOutput(content, cmd)
			if err != nil {
				fmt.Printf("❌ Failed to save output: %v\n", err)
				return nil
			}

			fmt.Printf("✅ Schema snapshot created successfully\n")
			if outputPath := cmd.Flag("output").Value.String(); outputPath != "" {
				fmt.Printf("📁 Saved to: %s\n", outputPath)
			}

			return nil
		},
	}
}
