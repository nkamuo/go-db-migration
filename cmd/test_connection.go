package cmd

import (
	"fmt"

	"github.com/nkamuo/go-db-migration/internal/database"
	"github.com/spf13/cobra"
)

// testConnectionCmd represents the test-connection command
var testConnectionCmd = &cobra.Command{
	Use:   "test-connection",
	Short: "Test database connection",
	Long: `Tests the database connection using the specified connection configuration.
This is useful for validating configuration before running validation commands.`,

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

		fmt.Printf("Testing connection to %s:%d/%s as %s...\n",
			dbConfig.Host, dbConfig.Port, dbConfig.Database, dbConfig.Username)

		// Connect to database
		db, err := database.NewConnection(dbConfig)
		if err != nil {
			return fmt.Errorf("❌ Connection failed: %w", err)
		}
		defer db.Close()

		fmt.Println("✅ Connection successful!")

		if connectionName != "" {
			fmt.Printf("Using connection: %s\n", connectionName)
		} else {
			fmt.Println("Using default connection")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testConnectionCmd)
}
