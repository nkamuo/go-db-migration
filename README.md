# Database Migration Validation Tool

A robust Go program designed to identify potential issues that could prevent successful database migrations.

## Features

- **Foreign Key Validation**: Identifies records that would violate foreign key constraints during migration
- **NOT NULL Constraint Validation**: Finds null values in columns that will be made NOT NULL
- **Schema Comparison**: Compares current database schema with target schema
- **Schema File Comparison**: Compare two schema files directly without database connections
- **Schema Export**: Export complete database schema with vendor-specific data types and full metadata
- **Schema Snapshots**: Create simplified schema snapshots for version tracking and quick comparisons
- **Automated Fix Commands**: Fix foreign key violations and null value issues with remove or set-null/default actions
- **Validation Configuration**: Configurable validation behavior with options to ignore missing tables/columns
- **Dry-Run Mode**: Test fix operations safely before applying changes
- **Multiple Output Formats**: Supports table, JSON, YAML, and CSV output formats
- **Flexible Configuration**: Supports multiple database connections with fallback to defaults


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

### Validation Configuration

Configure validation behavior by adding a `validation` section to your `conf.json`:

```json
{
    "DB": {
        "default": {
            "host": "localhost",
            "port": 5432,
            "username": "user",
            "password": "password",
            "database": "mydb"
        }
    },
    "validation": {
        "ignoreMissingTables": false,
        "ignoreMissingColumns": false,
        "stopOnFirstError": false,
        "maxIssuesPerTable": 100
    }
}
```

**Validation Options:**
- `ignoreMissingTables`: Skip validation for tables that don't exist in the database
- `ignoreMissingColumns`: Skip validation for columns that don't exist in the database  
- `stopOnFirstError`: Stop validation on the first error encountered
- `maxIssuesPerTable`: Maximum number of issues to report per table (prevents overwhelming output)

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

# Check NOT NULL constraints with validation options
./bin/migrator validate null --ignore-missing-tables --max-issues 10

# Fix foreign key violations by removing invalid records (dry-run first)
./bin/migrator fix fk --action remove --dry-run

# Fix foreign key violations by setting to null (with confirmation)
./bin/migrator fix fk --action set-null --confirm

# Fix null value violations by removing records
./bin/migrator fix null --action remove --dry-run

# Fix null value violations by setting to default values
./bin/migrator fix null --action set-default --confirm

# Compare current database vs target schema file
./bin/migrator schema compare

# Compare two schema files directly (no database connection needed)
./bin/migrator schema diff schema-v1.json schema-v2.json
./bin/migrator schema diff --source current.json --target new.json --format json

# Run all validations
./bin/migrator validate all

# Show schema information
./bin/migrator schema info

# Export complete database schema with full metadata
./bin/migrator schema export --format json -o database-schema.json

# Create simplified schema snapshot for version tracking
./bin/migrator schema snapshot --format json -o schema-snapshot.json
```

### Command Line Options

- `--config, -c`: Specify config file path (default: ./conf.json)
- `--connection`: Use named connection from config
- `--schema, -s`: Specify target schema file (default: ./schema.json)
- `--format, -f`: Output format (table, json, yaml, csv)
- `--output, -o`: Save output to file instead of stdout

#### Validation Options
- `--ignore-missing-tables`: Skip validation for tables that don't exist in database
- `--ignore-missing-columns`: Skip validation for columns that don't exist in database
- `--stop-on-error`: Stop validation on first error (default: continue)
- `--max-issues`: Maximum number of issues to report per table (default: 100)

#### Fix Command Options
- `--action`: Action to take (remove, set-null, set-default)
- `--dry-run`: Show what would be fixed without making changes (default: true)
- `--confirm`: Actually perform the fixes (required for real changes)

### Examples

```bash
# Use specific connection and output to file
./bin/migrator validate all --connection "JAMES Database" --output report.json --format json

# Validate foreign keys with YAML output
./bin/migrator validate fk --format yaml

# Validate null constraints with configuration options
./bin/migrator validate null --ignore-missing-tables --max-issues 5

# Fix foreign key violations safely (dry-run first)
./bin/migrator fix fk --action remove --dry-run --connection production

# Actually fix foreign key violations after reviewing dry-run
./bin/migrator fix fk --action set-null --confirm --connection production

# Fix null value issues by setting to defaults
./bin/migrator fix null --action set-default --dry-run

# Compare schemas and save as CSV
./bin/migrator schema compare --format csv --output schema-diff.csv

# Test connection to specific database
./bin/migrator connection test --connection "JAMES Database"

# Display schema information in JSON format
./bin/migrator schema info --format json

# Get help for specific command
./bin/migrator validate --help
```

### Schema Commands

The tool provides several schema-related commands for different use cases:

#### `schema compare`
Compares a live database schema with a target schema file. Requires a database connection.
```bash
# Compare current database with target schema file
./bin/migrator schema compare --connection production --schema target-schema.json
```

#### `schema diff`
Compares two schema files directly without requiring database connections. Perfect for:
- Comparing different versions of schema files
- Validating schema changes before deployment
- Code review processes

```bash
# Compare two schema files using positional arguments
./bin/migrator schema diff schema-v1.json schema-v2.json

# Compare using flags with custom output
./bin/migrator schema diff --source current.json --target new.json --format json -o diff.json
```

#### `schema export`
Exports the complete database schema with full metadata and vendor-specific data types.
```bash
# Export full schema with PostgreSQL-specific types like "character varying(100)"
./bin/migrator schema export --format json -o full-schema.json
```

#### `schema snapshot`
Creates simplified schema snapshots optimized for version tracking and quick comparisons.
```bash
# Create lightweight snapshot for version control
./bin/migrator schema snapshot --format json -o schema-v1.0.0.json
```

#### `schema validate`
Validates schema file structure for consistency and correctness.
```bash
# Check schema file for structural issues
./bin/migrator schema validate --schema my-schema.json
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
- Supports configuration options to handle missing tables/columns gracefully

### Schema Comparison
- Compares table structures between current and target schemas
- Identifies missing, extra, or modified tables and columns
- Highlights foreign key differences
- Detects data type and constraint changes

## Fix Commands

The tool provides automated fix commands to resolve data issues found during validation. All fix commands support dry-run mode for safe testing.

### Foreign Key Fixes

Fix foreign key violations using different strategies:

```bash
# Remove records with invalid foreign keys (dry-run)
./bin/migrator fix fk --action remove --dry-run

# Set foreign key columns to NULL (with confirmation)
./bin/migrator fix fk --action set-null --confirm

# Specify connection and output format
./bin/migrator fix fk --action remove --confirm --connection production --format json
```

**Actions:**
- `remove`: Delete records that have invalid foreign key references
- `set-null`: Set foreign key columns to NULL (only for nullable columns)

### Null Value Fixes

Fix null value violations in NOT NULL columns:

```bash
# Remove records with null values (dry-run)
./bin/migrator fix null --action remove --dry-run

# Set null values to column defaults (with confirmation)
./bin/migrator fix null --action set-default --confirm

# Check what would be fixed without making changes
./bin/migrator fix null --action set-default --dry-run --format json
```

**Actions:**
- `remove`: Delete records that have null values in NOT NULL columns
- `set-default`: Set null values to column default values (where defaults exist)

### Safety Features

- **Dry-run by default**: All fix commands run in dry-run mode unless `--confirm` is specified
- **Confirmation required**: Real changes require explicit `--confirm` flag
- **Detailed reporting**: Shows exactly what will be changed before and after
- **Transaction safety**: All fixes run within database transactions
- **Rollback capability**: Failed operations are automatically rolled back

### Fix Command Examples

```bash
# Safe workflow: validate, dry-run, then fix
./bin/migrator validate fk --format json > fk-issues.json
./bin/migrator fix fk --action remove --dry-run
./bin/migrator fix fk --action remove --confirm

# Fix null values with defaults where possible
./bin/migrator validate null --max-issues 10
./bin/migrator fix null --action set-default --dry-run
./bin/migrator fix null --action set-default --confirm

# Handle missing tables gracefully during validation
./bin/migrator validate null --ignore-missing-tables --ignore-missing-columns
```

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
│   ├── cli/              # CLI command implementations
│   │   ├── fix.go        # Fix command for database issues
│   │   └── validate.go   # Validation commands
│   ├── config/           # Configuration management
│   ├── database/         # Database operations and fixes
│   ├── models/           # Data models and schemas
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
./bin/migrator fix --help

# Get help for specific fix commands
./bin/migrator fix fk --help
./bin/migrator fix null --help
```

## License

[Add your license information here]
