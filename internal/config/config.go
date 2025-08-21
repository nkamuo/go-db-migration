package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DBConfig represents a database configuration
type DBConfig struct {
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Port     int    `json:"port" yaml:"port" mapstructure:"port"`
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
	Database string `json:"database" yaml:"database" mapstructure:"database"`
}

// Connection represents a named database connection
type Connection struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`
	Host     string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port"`
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`
	Database string `json:"database,omitempty" yaml:"database,omitempty" mapstructure:"database"`
}

// Config represents the main configuration structure
type Config struct {
	DB struct {
		Default     DBConfig     `json:"default" yaml:"default" mapstructure:"default"`
		Connections []Connection `json:"connections" yaml:"connections" mapstructure:"connections"`
	} `json:"DB" yaml:"DB" mapstructure:"DB"`
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
			return &config, nil
		}
	}

	return nil, fmt.Errorf("connection '%s' not found in configuration", connectionName)
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
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
		v.AddConfigPath("$HOME/.stsx-db-migration")
		v.AddConfigPath("/etc/stsx-db-migration")
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

// GetDefaultSchemaPath returns the default path for the schema file
func GetDefaultSchemaPath() string {
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "schema.json")
}
