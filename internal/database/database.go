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
	conn    *sql.DB
	config  *config.DBConfig
	dbType  DatabaseType
	dialect DatabaseDialect
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

// Ping tests the database connection
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// GetTableList retrieves just the table names without full schema
func (db *DB) GetTableList() ([]string, error) {
	return db.getTables()
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
		var charMaxLength sql.NullInt64
		var numericPrecision sql.NullInt64
		var numericScale sql.NullInt64
		var datetimePrecision sql.NullInt64

		if err := rows.Scan(
			&column.ColumnName,
			&column.DataType,
			&defaultValue,
			&column.IsNullable,
			&charMaxLength,
			&numericPrecision,
			&numericScale,
			&datetimePrecision,
		); err != nil {
			return nil, err
		}

		if defaultValue.Valid {
			column.DefaultValue = defaultValue.String
		}

		if charMaxLength.Valid {
			val := int(charMaxLength.Int64)
			column.CharacterMaxLength = &val
		}

		if numericPrecision.Valid {
			val := int(numericPrecision.Int64)
			column.NumericPrecision = &val
		}

		if numericScale.Valid {
			val := int(numericScale.Int64)
			column.NumericScale = &val
		}

		if datetimePrecision.Valid {
			val := int(datetimePrecision.Int64)
			column.DatetimePrecision = &val
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
			// Ensure the foreign key has the table name set (it might not be in the JSON)
			if fk.TableName == "" {
				fk.TableName = table.TableName
			}

			violations, err := db.findForeignKeyViolations(fk)
			if err != nil {
				// Instead of returning immediately, create a validation issue for the error
				issue := models.ValidationIssue{
					Type:     "foreign_key_validation_error",
					Severity: "error",
					Table:    fk.TableName,
					Column:   fk.ColumnName,
					Message:  fmt.Sprintf("Failed to validate foreign key '%s': %v", fk.ConstraintName, err),
					Details: map[string]interface{}{
						"constraint_name":   fk.ConstraintName,
						"referenced_table":  fk.ReferencedTable,
						"referenced_column": fk.ReferencedColumn,
						"error_type":        "validation_error",
					},
				}
				issues = append(issues, issue)
				continue // Continue to next foreign key instead of stopping
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
		// Create a validation issue instead of returning an error
		issue := models.ValidationIssue{
			Type:     "missing_source_table",
			Severity: "error",
			Table:    fk.TableName,
			Column:   fk.ColumnName,
			Message:  fmt.Sprintf("Source table '%s' does not exist in the database (required by foreign key constraint '%s')", fk.TableName, fk.ConstraintName),
			Details: map[string]interface{}{
				"constraint_name":   fk.ConstraintName,
				"referenced_table":  fk.ReferencedTable,
				"referenced_column": fk.ReferencedColumn,
				"error_type":        "missing_source_table",
			},
		}
		return []models.ValidationIssue{issue}, nil
	}

	referencedExists, err := db.tableExists(fk.ReferencedTable)
	if err != nil {
		return nil, fmt.Errorf("failed to check if referenced table '%s' exists: %w", fk.ReferencedTable, err)
	}
	if !referencedExists {
		// Create a validation issue instead of returning an error
		issue := models.ValidationIssue{
			Type:     "missing_referenced_table",
			Severity: "error",
			Table:    fk.TableName,
			Column:   fk.ColumnName,
			Message:  fmt.Sprintf("Referenced table '%s' does not exist in the database (required by foreign key constraint '%s')", fk.ReferencedTable, fk.ConstraintName),
			Details: map[string]interface{}{
				"constraint_name":   fk.ConstraintName,
				"referenced_table":  fk.ReferencedTable,
				"referenced_column": fk.ReferencedColumn,
				"error_type":        "missing_referenced_table",
			},
		}
		return []models.ValidationIssue{issue}, nil
	}

	// Check if the source column exists
	sourceColExists, err := db.columnExists(fk.TableName, fk.ColumnName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if source column '%s.%s' exists: %w", fk.TableName, fk.ColumnName, err)
	}
	if !sourceColExists {
		issue := models.ValidationIssue{
			Type:     "missing_source_column",
			Severity: "error",
			Table:    fk.TableName,
			Column:   fk.ColumnName,
			Message:  fmt.Sprintf("Source column '%s.%s' does not exist in the database (required by foreign key constraint '%s')", fk.TableName, fk.ColumnName, fk.ConstraintName),
			Details: map[string]interface{}{
				"constraint_name":   fk.ConstraintName,
				"referenced_table":  fk.ReferencedTable,
				"referenced_column": fk.ReferencedColumn,
				"error_type":        "missing_source_column",
			},
		}
		return []models.ValidationIssue{issue}, nil
	}

	// Check if the referenced column exists
	refColExists, err := db.columnExists(fk.ReferencedTable, fk.ReferencedColumn)
	if err != nil {
		return nil, fmt.Errorf("failed to check if referenced column '%s.%s' exists: %w", fk.ReferencedTable, fk.ReferencedColumn, err)
	}
	if !refColExists {
		issue := models.ValidationIssue{
			Type:     "missing_referenced_column",
			Severity: "error",
			Table:    fk.TableName,
			Column:   fk.ColumnName,
			Message:  fmt.Sprintf("Referenced column '%s.%s' does not exist in the database (required by foreign key constraint '%s')", fk.ReferencedTable, fk.ReferencedColumn, fk.ConstraintName),
			Details: map[string]interface{}{
				"constraint_name":   fk.ConstraintName,
				"referenced_table":  fk.ReferencedTable,
				"referenced_column": fk.ReferencedColumn,
				"error_type":        "missing_referenced_column",
			},
		}
		return []models.ValidationIssue{issue}, nil
	}

	// Build query to find orphaned records using dialect
	identifierCol := db.getIdentifierColumn(fk.TableName)
	query := db.dialect.GetForeignKeyViolationsQuery(fk, identifierCol)

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
	return db.ValidateNotNullConstraintsWithConfig(targetSchema, nil)
}

// ValidateNotNullConstraintsWithConfig checks for null values with validation configuration
func (db *DB) ValidateNotNullConstraintsWithConfig(targetSchema models.Schema, validationConfig *config.ValidationConfig) ([]models.ValidationIssue, error) {
	var issues []models.ValidationIssue
	var defaultConfig config.ValidationConfig

	if validationConfig == nil {
		defaultConfig = config.ValidationConfig{
			IgnoreMissingTables:  false,
			IgnoreMissingColumns: false,
			StopOnFirstError:     false,
			MaxIssuesPerTable:    1000,
		}
		validationConfig = &defaultConfig
	}

	for _, table := range targetSchema {
		// Check if table exists
		tableExists, err := db.tableExists(table.TableName)
		if err != nil {
			if validationConfig.StopOnFirstError {
				return nil, fmt.Errorf("failed to check if table %s exists: %w", table.TableName, err)
			}
			// Add as validation issue and continue
			issues = append(issues, models.ValidationIssue{
				Type:     "table_check_error",
				Severity: "error",
				Table:    table.TableName,
				Message:  fmt.Sprintf("Failed to check if table exists: %v", err),
			})
			continue
		}

		if !tableExists {
			if validationConfig.IgnoreMissingTables {
				continue // Skip this table
			}
			// Add as validation issue
			issues = append(issues, models.ValidationIssue{
				Type:     "missing_table",
				Severity: "warning",
				Table:    table.TableName,
				Message:  fmt.Sprintf("Table '%s' does not exist in database", table.TableName),
			})
			continue
		}

		for _, column := range table.Columns {
			if column.IsNotNull() {
				// Check if column exists
				columnExists, err := db.columnExists(table.TableName, column.ColumnName)
				if err != nil {
					if validationConfig.StopOnFirstError {
						return nil, fmt.Errorf("failed to check if column %s.%s exists: %w", table.TableName, column.ColumnName, err)
					}
					issues = append(issues, models.ValidationIssue{
						Type:     "column_check_error",
						Severity: "error",
						Table:    table.TableName,
						Column:   column.ColumnName,
						Message:  fmt.Sprintf("Failed to check if column exists: %v", err),
					})
					continue
				}

				if !columnExists {
					if validationConfig.IgnoreMissingColumns {
						continue // Skip this column
					}
					issues = append(issues, models.ValidationIssue{
						Type:     "missing_column",
						Severity: "warning",
						Table:    table.TableName,
						Column:   column.ColumnName,
						Message:  fmt.Sprintf("Column '%s.%s' does not exist in database", table.TableName, column.ColumnName),
					})
					continue
				}

				violations, err := db.findNullViolations(table.TableName, column, validationConfig.MaxIssuesPerTable)
				if err != nil {
					if validationConfig.StopOnFirstError {
						return nil, fmt.Errorf("failed to validate NOT NULL constraint for %s.%s: %w", table.TableName, column.ColumnName, err)
					}
					issues = append(issues, models.ValidationIssue{
						Type:     "validation_error",
						Severity: "error",
						Table:    table.TableName,
						Column:   column.ColumnName,
						Message:  fmt.Sprintf("Failed to validate NOT NULL constraint: %v", err),
					})
					continue
				}
				issues = append(issues, violations...)
			}
		}
	}

	return issues, nil
}

// findNullViolations finds records with null values in columns that should be NOT NULL
func (db *DB) findNullViolations(tableName string, column models.Column, maxIssues ...int) ([]models.ValidationIssue, error) {
	limit := 1000 // Default limit
	if len(maxIssues) > 0 && maxIssues[0] > 0 {
		limit = maxIssues[0]
	}

	identifierCol := db.getIdentifierColumn(tableName)

	// Build query with custom limit (using quoted identifiers for PostgreSQL compatibility)
	query := fmt.Sprintf(`
		SELECT "%s"
		FROM "%s"
		WHERE "%s" IS NULL
		LIMIT %d`, identifierCol, tableName, column.ColumnName, limit)

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
		exists, err := db.columnExists(tableName, pk)
		if err == nil && exists {
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
func (db *DB) columnExists(tableName, columnName string) (bool, error) {
	query := db.dialect.GetColumnExistsQuery()
	var exists int
	err := db.conn.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetTableRowCount returns the number of rows in a table
func (db *DB) GetTableRowCount(tableName string) (int64, error) {
	query := db.dialect.GetTableRowCountQuery(tableName)
	var count int64
	err := db.conn.QueryRow(query).Scan(&count)
	return count, err
}

// FixForeignKeyViolations fixes foreign key constraint violations
func (db *DB) FixForeignKeyViolations(targetSchema models.Schema, action string, dryRun bool, validationConfig *config.ValidationConfig) (models.FixResults, error) {
	results := make(models.FixResults)

	for _, table := range targetSchema {
		for _, fk := range table.ForeignKeys {
			// Ensure the foreign key has the table name set (it might not be in the JSON)
			if fk.TableName == "" {
				fk.TableName = table.TableName
			}
			
			tableName := fk.TableName
			if _, exists := results[tableName]; !exists {
				results[tableName] = models.FixResult{}
			}

			// Find violations first
			violations, err := db.findForeignKeyViolations(fk)
			if err != nil {
				if validationConfig != nil && validationConfig.IgnoreMissingTables {
					continue
				}
				result := results[tableName]
				result.Error = err.Error()
				results[tableName] = result
				continue
			}

			violationCount := len(violations)
			if violationCount == 0 {
				continue
			}

			result := results[tableName]
			result.IssuesFound += violationCount

			if !dryRun {
				var recordsAffected int
				var fixErr error

				switch action {
				case "remove":
					recordsAffected, fixErr = db.removeForeignKeyViolatingRecords(fk)
				case "set-null":
					recordsAffected, fixErr = db.setForeignKeyColumnsToNull(fk)
				default:
					fixErr = fmt.Errorf("unknown action: %s", action)
				}

				if fixErr != nil {
					result.Error = fixErr.Error()
					result.Success = false
				} else {
					result.RecordsAffected += recordsAffected
					result.Success = true
				}
			} else {
				// In dry-run mode, count what would be affected
				result.RecordsAffected += violationCount
				result.Success = true
				result.Details = fmt.Sprintf("Would %s %d records", action, violationCount)
			}

			results[tableName] = result
		}
	}

	return results, nil
}

// FixNullValueViolations fixes NULL value violations for NOT NULL constraints
func (db *DB) FixNullValueViolations(targetSchema models.Schema, action, defaultValue string, dryRun bool, validationConfig *config.ValidationConfig) (models.FixResults, error) {
	results := make(models.FixResults)

	for _, table := range targetSchema {
		tableName := table.TableName
		if _, exists := results[tableName]; !exists {
			results[tableName] = models.FixResult{}
		}

		for _, column := range table.Columns {
			if !column.IsNotNull() {
				continue
			}

			// Check if table/column exists
			if validationConfig != nil {
				if validationConfig.IgnoreMissingTables {
					tableExists, _ := db.tableExists(tableName)
					if !tableExists {
						continue
					}
				}
				if validationConfig.IgnoreMissingColumns {
					columnExists, _ := db.columnExists(tableName, column.ColumnName)
					if !columnExists {
						continue
					}
				}
			}

			// Find null violations
			violations, err := db.findNullViolations(tableName, column)
			if err != nil {
				result := results[tableName]
				result.Error = err.Error()
				results[tableName] = result
				continue
			}

			violationCount := len(violations)
			if violationCount == 0 {
				continue
			}

			result := results[tableName]
			result.IssuesFound += violationCount

			if !dryRun {
				var recordsAffected int
				var fixErr error

				switch action {
				case "remove":
					recordsAffected, fixErr = db.removeNullValueRecords(tableName, column.ColumnName)
				case "set-default":
					recordsAffected, fixErr = db.setNullValuesToDefault(tableName, column.ColumnName, defaultValue)
				default:
					fixErr = fmt.Errorf("unknown action: %s", action)
				}

				if fixErr != nil {
					result.Error = fixErr.Error()
					result.Success = false
				} else {
					result.RecordsAffected += recordsAffected
					result.Success = true
				}
			} else {
				// In dry-run mode, count what would be affected
				result.RecordsAffected += violationCount
				result.Success = true
				result.Details = fmt.Sprintf("Would %s %d records in column %s", action, violationCount, column.ColumnName)
			}

			results[tableName] = result
		}
	}

	return results, nil
}

// Helper methods for actual fix operations

func (db *DB) removeForeignKeyViolatingRecords(fk models.ForeignKey) (int, error) {
	query := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE "%s" IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM "%s" AS ref_table
			WHERE ref_table."%s" = "%s"."%s"
		  )`,
		fk.TableName,
		fk.ColumnName,
		fk.ReferencedTable,
		fk.ReferencedColumn,
		fk.TableName,
		fk.ColumnName)

	result, err := db.conn.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

func (db *DB) setForeignKeyColumnsToNull(fk models.ForeignKey) (int, error) {
	query := fmt.Sprintf(`
		UPDATE "%s"
		SET "%s" = NULL
		WHERE "%s" IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM "%s" AS ref_table
			WHERE ref_table."%s" = "%s"."%s"
		  )`,
		fk.TableName,
		fk.ColumnName,
		fk.ColumnName,
		fk.ReferencedTable,
		fk.ReferencedColumn,
		fk.TableName,
		fk.ColumnName)

	result, err := db.conn.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

func (db *DB) removeNullValueRecords(tableName, columnName string) (int, error) {
	query := fmt.Sprintf(`
		DELETE FROM "%s"
		WHERE "%s" IS NULL`,
		tableName,
		columnName)

	result, err := db.conn.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}

func (db *DB) setNullValuesToDefault(tableName, columnName, defaultValue string) (int, error) {
	query := fmt.Sprintf(`
		UPDATE "%s"
		SET "%s" = $1
		WHERE "%s" IS NULL`,
		tableName,
		columnName,
		columnName)

	result, err := db.conn.Exec(query, defaultValue)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return int(rowsAffected), err
}
