package database

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	
	"github.com/nkamuo/go-db-migration/internal/config"
	"github.com/nkamuo/go-db-migration/internal/models"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	MySQL      DatabaseType = "mysql"
)

// DB represents a database connection wrapper with multi-vendor support
type DB struct {
	conn     *sql.DB
	config   *config.DBConfig
	dbType   DatabaseType
	dialect  DatabaseDialect
}

// DatabaseDialect interface for vendor-specific SQL queries
type DatabaseDialect interface {
	GetTablesQuery() string
	GetColumnsQuery() string
	GetForeignKeysQuery() string
	GetColumnExistsQuery() string
	BuildConnectionString(cfg *config.DBConfig) string
	GetDriverName() string
	GetIdentifierQuote() string
	GetTableRowCountQuery(tableName string) string
	GetNullViolationsQuery(tableName, columnName, identifierCol string) string
	GetForeignKeyViolationsQuery(fk models.ForeignKey, identifierCol string) string
}

// NewConnection creates a new database connection with the appropriate dialect
func NewConnection(cfg *config.DBConfig) (*DB, error) {
	dbType := DatabaseType(cfg.Type)
	if dbType == "" {
		dbType = PostgreSQL // Default to PostgreSQL
	}

	var dialect DatabaseDialect
	switch dbType {
	case PostgreSQL:
		dialect = &PostgreSQLDialect{}
	case MySQL:
		dialect = &MySQLDialect{}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	connStr := dialect.BuildConnectionString(cfg)
	conn, err := sql.Open(dialect.GetDriverName(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		conn:    conn,
		config:  cfg,
		dbType:  dbType,
		dialect: dialect,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetDatabaseType returns the database type
func (db *DB) GetDatabaseType() DatabaseType {
	return db.dbType
}

// GetCurrentSchema retrieves the current database schema
func (db *DB) GetCurrentSchema() (models.Schema, error) {
	tables, err := db.getTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	var schema models.Schema
	for _, tableName := range tables {
		table := models.Table{TableName: tableName}

		// Get columns
		columns, err := db.getTableColumns(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		table.Columns = columns

		// Get foreign keys
		foreignKeys, err := db.getTableForeignKeys(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get foreign keys for table %s: %w", tableName, err)
		}
		table.ForeignKeys = foreignKeys

		schema = append(schema, table)
	}

	return schema, nil
}

// getTables retrieves all table names from the database
func (db *DB) getTables() ([]string, error) {
	query := db.dialect.GetTablesQuery()
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// getTableColumns retrieves all columns for a specific table
func (db *DB) getTableColumns(tableName string) ([]models.Column, error) {
	query := db.dialect.GetColumnsQuery()
	rows, err := db.conn.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.Column
	for rows.Next() {
		var column models.Column
		var defaultValue sql.NullString

		if err := rows.Scan(
			&column.ColumnName,
			&column.DataType,
			&defaultValue,
			&column.IsNullable,
		); err != nil {
			return nil, err
		}

		if defaultValue.Valid {
			column.DefaultValue = defaultValue.String
		}

		columns = append(columns, column)
	}

	return columns, rows.Err()
}

// getTableForeignKeys retrieves all foreign keys for a specific table
func (db *DB) getTableForeignKeys(tableName string) ([]models.ForeignKey, error) {
	query := db.dialect.GetForeignKeysQuery()
	rows, err := db.conn.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var foreignKeys []models.ForeignKey
	for rows.Next() {
		var fk models.ForeignKey
		if err := rows.Scan(
			&fk.ConstraintName,
			&fk.TableName,
			&fk.ColumnName,
			&fk.ReferencedTable,
			&fk.ReferencedColumn,
			&fk.UpdateRule,
			&fk.DeleteRule,
		); err != nil {
			return nil, err
		}
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// tableExists checks if a table exists in the database
func (db *DB) tableExists(tableName string) (bool, error) {
	query := `
		SELECT 1 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		  AND table_name = $1`
	
	var exists int
	err := db.conn.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ValidateForeignKeys checks for foreign key constraint violations
func (db *DB) ValidateForeignKeys(targetSchema models.Schema) ([]models.ValidationIssue, error) {
	var issues []models.ValidationIssue

	for _, table := range targetSchema {
		for _, fk := range table.ForeignKeys {
			violations, err := db.findForeignKeyViolations(fk)
			if err != nil {
				return nil, fmt.Errorf("failed to validate foreign key %s: %w", fk.ConstraintName, err)
			}
			issues = append(issues, violations...)
		}
	}

	return issues, nil
}

// findForeignKeyViolations finds records that violate a foreign key constraint
func (db *DB) findForeignKeyViolations(fk models.ForeignKey) ([]models.ValidationIssue, error) {
	// First, check if both tables exist
	sourceExists, err := db.tableExists(fk.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if source table '%s' exists: %w", fk.TableName, err)
	}
	if !sourceExists {
		return nil, fmt.Errorf("source table '%s' does not exist in the database", fk.TableName)
	}

	referencedExists, err := db.tableExists(fk.ReferencedTable)
	if err != nil {
		return nil, fmt.Errorf("failed to check if referenced table '%s' exists: %w", fk.ReferencedTable, err)
	}
	if !referencedExists {
		return nil, fmt.Errorf("referenced table '%s' does not exist in the database (required by foreign key constraint '%s')", fk.ReferencedTable, fk.ConstraintName)
	}

	// Build query to find orphaned records
	query := fmt.Sprintf(`
		SELECT %s, %s
		FROM %s AS source_table
		WHERE %s IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM %s AS ref_table
			WHERE ref_table.%s = source_table.%s
		  )
		LIMIT 1000`, // Limit to prevent overwhelming output
		fk.ColumnName,
		db.getIdentifierColumn(fk.TableName),
		fk.TableName,
		fk.ColumnName,
		fk.ReferencedTable,
		fk.ReferencedColumn,
		fk.ColumnName)

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute foreign key validation query for constraint '%s' (table: %s, column: %s, references: %s.%s): %w", 
			fk.ConstraintName, fk.TableName, fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn, err)
	}
	defer rows.Close()

	var issues []models.ValidationIssue
	for rows.Next() {
		var foreignKeyValue, identifier sql.NullString
		if err := rows.Scan(&foreignKeyValue, &identifier); err != nil {
			return nil, err
		}

		issue := models.ValidationIssue{
			Type:     "foreign_key_violation",
			Severity: "error",
			Table:    fk.TableName,
			Column:   fk.ColumnName,
			Message: fmt.Sprintf("Foreign key violation: value '%s' references non-existent record in %s.%s",
				foreignKeyValue.String, fk.ReferencedTable, fk.ReferencedColumn),
			PrimaryKey: foreignKeyValue.String,
			Identifier: identifier.String,
			Details: map[string]interface{}{
				"constraint_name":   fk.ConstraintName,
				"referenced_table":  fk.ReferencedTable,
				"referenced_column": fk.ReferencedColumn,
				"foreign_key_value": foreignKeyValue.String,
			},
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

// ValidateNotNullConstraints checks for null values in columns that should be NOT NULL
func (db *DB) ValidateNotNullConstraints(targetSchema models.Schema) ([]models.ValidationIssue, error) {
	var issues []models.ValidationIssue

	for _, table := range targetSchema {
		for _, column := range table.Columns {
			if column.IsNotNull() {
				violations, err := db.findNullViolations(table.TableName, column)
				if err != nil {
					return nil, fmt.Errorf("failed to validate NOT NULL constraint for %s.%s: %w",
						table.TableName, column.ColumnName, err)
				}
				issues = append(issues, violations...)
			}
		}
	}

	return issues, nil
}

// findNullViolations finds records with null values in columns that should be NOT NULL
func (db *DB) findNullViolations(tableName string, column models.Column) ([]models.ValidationIssue, error) {
	identifierCol := db.getIdentifierColumn(tableName)

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s IS NULL
		LIMIT 1000`, // Limit to prevent overwhelming output
		identifierCol,
		tableName,
		column.ColumnName)

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []models.ValidationIssue
	for rows.Next() {
		var identifier sql.NullString
		if err := rows.Scan(&identifier); err != nil {
			return nil, err
		}

		issue := models.ValidationIssue{
			Type:     "null_constraint_violation",
			Severity: "error",
			Table:    tableName,
			Column:   column.ColumnName,
			Message: fmt.Sprintf("NULL value found in column '%s' which will be set to NOT NULL",
				column.ColumnName),
			Identifier: identifier.String,
			Details: map[string]interface{}{
				"data_type": column.DataType,
			},
		}
		issues = append(issues, issue)
	}

	return issues, rows.Err()
}

// getIdentifierColumn returns the best column to use as an identifier for a table
func (db *DB) getIdentifierColumn(tableName string) string {
	// Try common primary key patterns
	possiblePKs := []string{
		"id",
		tableName + "_id",
		"uuid",
		"guid",
		"key",
	}

	for _, pk := range possiblePKs {
		if db.columnExists(tableName, pk) {
			return pk
		}
	}

	// Fall back to first column
	columns, err := db.getTableColumns(tableName)
	if err == nil && len(columns) > 0 {
		return columns[0].ColumnName
	}

	return "1" // Fallback to literal
}

// columnExists checks if a column exists in a table
func (db *DB) columnExists(tableName, columnName string) bool {
	query := db.dialect.GetColumnExistsQuery()
	var exists int
	err := db.conn.QueryRow(query, tableName, columnName).Scan(&exists)
	return err == nil
}

// GetTableRowCount returns the number of rows in a table
func (db *DB) GetTableRowCount(tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	var count int64
	err := db.conn.QueryRow(query).Scan(&count)
	return count, err
}
