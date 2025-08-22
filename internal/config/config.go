package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DBConfig represents a database configuration
type DBConfig struct {
	Type     string `json:"type" yaml:"type" mapstructure:"type"` // postgres, mysql
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	Database string `json:"database" yaml:"database" mapstructure:"database"`
	SSLMode  string `json:"sslmode,omitempty" yaml:"sslmode,omitempty" mapstructure:"sslmode"` // For PostgreSQL
}

// ValidationConfig represents validation behavior configuration
type ValidationConfig struct {
	IgnoreMissingTables  bool `json:"ignore_missing_tables" yaml:"ignore_missing_tables" mapstructure:"ignore_missing_tables"`
	IgnoreMissingColumns bool `json:"ignore_missing_columns" yaml:"ignore_missing_columns" mapstructure:"ignore_missing_columns"`
	StopOnFirstError     bool `json:"stop_on_first_error" yaml:"stop_on_first_error" mapstructure:"stop_on_first_error"`
	MaxIssuesPerTable    int  `json:"max_issues_per_table" yaml:"max_issues_per_table" mapstructure:"max_issues_per_table"`
}

// Connection represents a named database connection
type Connection struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type"`
	Host     string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port"`
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`
	Database string `json:"database,omitempty" yaml:"database,omitempty" mapstructure:"database"`
	SSLMode  string `json:"sslmode,omitempty" yaml:"sslmode,omitempty" mapstructure:"sslmode"`
}

// Config represents the main configuration structure
type Config struct {
	DB struct {
		Default     DBConfig     `json:"default" yaml:"default" mapstructure:"default"`
		Connections []Connection `json:"connections" yaml:"connections" mapstructure:"connections"`
	} `json:"DB" yaml:"DB" mapstructure:"DB"`
	Validation ValidationConfig `json:"validation" yaml:"validation" mapstructure:"validation"`
}

// GetConnectionConfig returns the database configuration for a given connection name
// If the connection name is empty or not found, it returns the default configuration
func (c *Config) GetConnectionConfig(connectionName string) (*DBConfig, error) {
	if connectionName == "" {
		return &c.DB.Default, nil
	}

	// Find the named connection
	for _, conn := range c.DB.Connections {
		if conn.Name == connectionName {
			// Merge with defaults for any missing values
			config := c.DB.Default
			if conn.Type != "" {
				config.Type = conn.Type
			}
			if conn.Host != "" {
				config.Host = conn.Host
			}
			if conn.Port != 0 {
				config.Port = conn.Port
			}
			if conn.Username != "" {
				config.Username = conn.Username
			}
			if conn.Password != "" {
				config.Password = conn.Password
			}
			if conn.Database != "" {
				config.Database = conn.Database
			}
			if conn.SSLMode != "" {
				config.SSLMode = conn.SSLMode
			}
			return &config, nil
		}
	}

	return nil, fmt.Errorf("connection '%s' not found in configuration", connectionName)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate default database type
	if c.DB.Default.Type == "" {
		c.DB.Default.Type = "postgres" // Default to postgres
	}
	if c.DB.Default.Type != "postgres" && c.DB.Default.Type != "mysql" {
		return fmt.Errorf("default database type must be 'postgres' or 'mysql', got '%s'", c.DB.Default.Type)
	}

	if c.DB.Default.Host == "" {
		return fmt.Errorf("default database host is required")
	}
	if c.DB.Default.Port <= 0 {
		return fmt.Errorf("default database port must be greater than 0")
	}
	if c.DB.Default.Username == "" {
		return fmt.Errorf("default database username is required")
	}
	if c.DB.Default.Database == "" {
		return fmt.Errorf("default database name is required")
	}

	// Validate connections
	for i, conn := range c.DB.Connections {
		if conn.Name == "" {
			return fmt.Errorf("connection at index %d must have a name", i)
		}
		// Validate connection type if specified
		if conn.Type != "" && conn.Type != "postgres" && conn.Type != "mysql" {
			return fmt.Errorf("connection '%s' has invalid type '%s', must be 'postgres' or 'mysql'", conn.Name, conn.Type)
		}
	}

	return nil
}

// LoadConfig loads configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	if configPath != "" {
		// Use specified config file
		v.SetConfigFile(configPath)
	} else {
		// Search for config file in current directory and common locations
		v.SetConfigName("conf")
		v.SetConfigType("json")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("$HOME/.migrator")
		v.AddConfigPath("/etc/migrator")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// GetValidationConfig returns the validation configuration with defaults
func (c *Config) GetValidationConfig() ValidationConfig {
	// Set defaults if not specified
	validationConfig := c.Validation
	if validationConfig.MaxIssuesPerTable == 0 {
		validationConfig.MaxIssuesPerTable = 1000 // Default limit
	}
	return validationConfig
}

// GetDefaultSchemaPath returns the default path for the schema file
func GetDefaultSchemaPath() string {
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "schema.json")
}
