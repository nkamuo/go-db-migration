package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/nkamuo/go-db-migration/internal/database"
)

// newConnectionCmd creates the connection command group
func newConnectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Database connection management commands",
		Long: `Commands for managing and testing database connections.
This includes testing connectivity, listing available connections,
and displaying connection information.`,
		Aliases: []string{"conn", "db"},
	}

	cmd.AddCommand(newConnectionTestCmd())
	cmd.AddCommand(newConnectionListCmd())
	cmd.AddCommand(newConnectionInfoCmd())

	return cmd
}

// newConnectionTestCmd creates the connection test command
func newConnectionTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test database connection",
		Long: `Tests the database connection to ensure it's working properly.
This command will attempt to connect to the database and perform
basic connectivity checks.`,
		Aliases: []string{"ping", "check"},

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

			connName := connectionName
			if connName == "" {
				connName = "default"
			}

			fmt.Printf("🔄 Testing connection '%s'...\n", connName)
			fmt.Printf("   Host: %s:%d\n", dbConfig.Host, dbConfig.Port)
			fmt.Printf("   Database: %s\n", dbConfig.Database)
			fmt.Printf("   User: %s\n", dbConfig.Username)

			// Test connection
			db, err := database.NewConnection(dbConfig)
			if err != nil {
				fmt.Printf("❌ Connection failed: %v\n", err)
				return fmt.Errorf("connection test failed")
			}
			defer db.Close()

			fmt.Printf("✅ Connection successful!\n")

			// Get some basic database info
			currentSchema, err := db.GetCurrentSchema()
			if err != nil {
				fmt.Printf("⚠️  Warning: Could not retrieve schema information: %v\n", err)
			} else {
				fmt.Printf("📊 Found %d tables in database\n", len(currentSchema))
			}

			return nil
		},
	}
}

// newConnectionListCmd creates the connection list command
func newConnectionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available database connections",
		Long: `Lists all available database connections defined in the configuration file.
This includes both the default connection and any named connections.`,
		Aliases: []string{"ls"},

		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := getConfigFromCmd(cmd)
			if err != nil {
				return err
			}

			fmt.Printf("📋 Available Database Connections\n")
			fmt.Printf("=================================\n\n")

			// Show default connection
			fmt.Printf("🏠 Default Connection:\n")
			fmt.Printf("   Host: %s:%d\n", cfg.DB.Default.Host, cfg.DB.Default.Port)
			fmt.Printf("   Database: %s\n", cfg.DB.Default.Database)
			fmt.Printf("   User: %s\n", cfg.DB.Default.Username)
			fmt.Printf("\n")

			// Show named connections
			if len(cfg.DB.Connections) > 0 {
				fmt.Printf("📝 Named Connections:\n")
				for i, conn := range cfg.DB.Connections {
					fmt.Printf("%d. %s\n", i+1, conn.Name)
					if conn.Database != "" {
						fmt.Printf("   Database: %s\n", conn.Database)
					}
					if conn.Host != "" {
						fmt.Printf("   Host: %s", conn.Host)
						if conn.Port != 0 {
							fmt.Printf(":%d", conn.Port)
						}
						fmt.Printf("\n")
					}
					if conn.Username != "" {
						fmt.Printf("   User: %s\n", conn.Username)
					}
					fmt.Printf("\n")
				}
			} else {
				fmt.Printf("📝 No named connections configured.\n\n")
			}

			fmt.Printf("💡 Use --connection <name> to specify a named connection\n")
			fmt.Printf("💡 Without --connection, the default connection will be used\n")

			return nil
		},
	}
}

// newConnectionInfoCmd creates the connection info command
func newConnectionInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display detailed connection information",
		Long: `Displays detailed information about a specific database connection
including resolved configuration values and connection status.`,
		Aliases: []string{"show"},

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

			connName := connectionName
			if connName == "" {
				connName = "default"
			}

			fmt.Printf("🔍 Connection Information: %s\n", connName)
			fmt.Printf("===============================\n\n")
			fmt.Printf("🏠 Host: %s\n", dbConfig.Host)
			fmt.Printf("🔌 Port: %d\n", dbConfig.Port)
			fmt.Printf("🗄️  Database: %s\n", dbConfig.Database)
			fmt.Printf("👤 Username: %s\n", dbConfig.Username)
			fmt.Printf("🔒 Password: %s\n", func() string {
				if dbConfig.Password != "" {
					return "[configured]"
				}
				return "[not set]"
			}())

			// Test connection and show status
			fmt.Printf("\n🔄 Testing connection...\n")
			db, err := database.NewConnection(dbConfig)
			if err != nil {
				fmt.Printf("❌ Status: Connection failed\n")
				fmt.Printf("💥 Error: %v\n", err)
				return nil
			}
			defer db.Close()

			fmt.Printf("✅ Status: Connected successfully\n")

			// Get additional database info
			currentSchema, err := db.GetCurrentSchema()
			if err == nil {
				fmt.Printf("📊 Tables: %d\n", len(currentSchema))
			}

			return nil
		},
	}
}
