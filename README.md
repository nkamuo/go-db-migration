# Database Migration Validation Tool

A robust Go program designed to identify potential issues that could prevent successful database migrations.

## Features

- **Foreign Key Validation**: Identifies records that would violate foreign key constraints during migration
- **NOT NULL Constraint Validation**: Finds null values in columns that will be made NOT NULL
- **Schema Comparison**: Compares current database schema with target schema
- **Multiple Output Formats**: Supports table, JSON, YAML, and CSV output formats
- **Flexible Configuration**: Supports multiple database connections with fallback to defaults
- **Enterprise-Ready**: Robust error handling, logging, and validation

## Installation

### Prerequisites

- Go 1.19 or later
- PostgreSQL database access
- Valid database credentials

### Building from Source

```bash
# Clone the repository
git clone https://github.com/nkamuo/go-db-migration.git
cd db-migration

# Install dependencies
make deps

# Build the binary
make build

# Or build for all platforms
make build-all
```

### Development Setup

```bash
# Build and prepare for development
make dev
```

## Configuration

The tool uses a `conf.json` file for database configuration:

```json
{
    "DB": {
        "default": {
            "host": "localhost",
            "port": 5432,
            "username": "user",
            "password": "password",
            "database": "mydb"
        },
        "connections": [
            {
                "name": "JAMES Database",
                "database": "james_db"
            }
        ]
    }
}
```

### Configuration Options

- **default**: Default database connection parameters
- **connections**: Named connections that inherit from default and override specific values

## Build

### Using Make (Recommended)

```bash
# Build the binary
make build

# Run tests
make test

# Clean build artifacts
make clean

# Install dependencies
make deps

# Build for multiple platforms
make build-all

# Show all available commands
make help
```

### Manual Build

```bash
# Build the migrator binary
go build -o bin/migrator ./cmd/migrator

# Or install directly
go install ./cmd/migrator
```

## Usage

### Basic Commands

```bash
# Test database connection
./bin/migrator connection test

# Validate schema file structure
./bin/migrator schema validate

# Check foreign key constraints
./bin/migrator validate fk

# Check NOT NULL constraints
./bin/migrator validate null

# Compare current vs target schema
./bin/migrator schema compare

# Run all validations
./bin/migrator validate all

# Show schema information
./bin/migrator schema info
```

### Command Line Options

- `--config, -c`: Specify config file path (default: ./conf.json)
- `--connection`: Use named connection from config
- `--schema, -s`: Specify target schema file (default: ./schema.json)
- `--format, -f`: Output format (table, json, yaml, csv)
- `--output, -o`: Save output to file instead of stdout

### Examples

```bash
# Use specific connection and output to file
./bin/migrator validate all --connection "JAMES Database" --output report.json --format json

# Validate foreign keys with YAML output
./bin/migrator validate fk --format yaml

# Compare schemas and save as CSV
./bin/migrator schema compare --format csv --output schema-diff.csv

# Test connection to specific database
./bin/migrator connection test --connection "JAMES Database"

# Display schema information in JSON format
./bin/migrator schema info --format json

# Get help for specific command
./bin/migrator validate --help
```

## Schema File Format

The target schema should be a JSON file with the following structure:

```json
[
    {
        "TableName": "users",
        "Columns": [
            {
                "ColumnName": "id",
                "DataType": "integer",
                "DefaultValue": null,
                "IsNullable": "NO"
            }
        ],
        "ForeignKeys": [
            {
                "ConstraintName": "fk_user_role",
                "TableName": "users",
                "ColumnName": "role_id",
                "ReferencedTable": "roles",
                "ReferencedColumn": "id",
                "UpdateRule": "CASCADE",
                "DeleteRule": "RESTRICT"
            }
        ]
    }
]
```

## Output Formats

### Table Format (Default)
Pretty-printed tables with color coding for easy reading in terminal.

### JSON Format
Structured JSON suitable for programmatic processing and integration.

### YAML Format
Human-readable YAML format for documentation and configuration management.

### CSV Format
Comma-separated values for spreadsheet analysis and reporting.

## Validation Types

### Foreign Key Validation
- Identifies orphaned records that reference non-existent parent records
- Provides primary key values and identifiers for easy record location
- Checks all foreign key constraints defined in target schema

### NOT NULL Constraint Validation
- Finds records with null values in columns marked as NOT NULL
- Helps identify data cleanup requirements before migration
- Provides record identifiers for targeted data fixes

### Schema Comparison
- Compares table structures between current and target schemas
- Identifies missing, extra, or modified tables and columns
- Highlights foreign key differences
- Detects data type and constraint changes

## Development

### Project Structure

```
.
├── cmd/                    # Command implementations
│   ├── root.go            # Root command and global flags
│   ├── validate_fk.go     # Foreign key validation
│   ├── validate_null.go   # NOT NULL validation
│   ├── compare_schema.go  # Schema comparison
│   ├── validate_all.go    # Comprehensive validation
│   ├── test_connection.go # Connection testing
│   └── version.go         # Version information
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── database/         # Database operations
│   ├── models/           # Data models
│   ├── output/           # Output formatting
│   └── schema/           # Schema operations
├── bin/                  # Built binaries
├── conf.json            # Configuration file
├── schema.json          # Target schema definition
├── main.go             # Application entry point
├── Makefile           # Build automation
└── README.md         # This file
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make format
```

### Building

```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all

# Clean build artifacts
make clean
```

## Error Handling

The tool provides comprehensive error handling with clear messages:

- Configuration validation errors
- Database connection issues
- Schema file parsing problems
- SQL execution errors
- Output formatting failures

## Best Practices

1. **Test Configuration**: Always run `test-connection` before validation commands
2. **Validate Schema**: Use `validate-schema` to check your schema file before running validations
3. **Start Small**: Run individual validation commands before `validate-all`
4. **Save Reports**: Use `--output` to save results for later analysis
5. **Use Appropriate Formats**: Choose output format based on your workflow (JSON for automation, table for review)

## Troubleshooting

### Common Issues

1. **Connection Failed**: Check database credentials and network connectivity
2. **Schema File Not Found**: Ensure schema file path is correct and file is readable
3. **Invalid Configuration**: Run config validation and check JSON syntax
4. **Permission Errors**: Ensure database user has SELECT permissions on all tables

### Getting Help

```bash
# Show all available commands
./bin/migrator help

# Get help for specific command
./bin/migrator validate fk --help

# Check version information
./bin/migrator version

# Show nested command help
./bin/migrator schema --help
./bin/migrator validate --help
./bin/migrator connection --help
```

## License

[Add your license information here]
