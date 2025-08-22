package models

import "fmt"

// Column represents a database column from the schema
type Column struct {
	ColumnName         string      `json:"ColumnName"`
	DataType           string      `json:"DataType"`
	DefaultValue       interface{} `json:"DefaultValue"`
	IsNullable         string      `json:"IsNullable"`
	CharacterMaxLength *int        `json:"CharacterMaxLength,omitempty"`
	NumericPrecision   *int        `json:"NumericPrecision,omitempty"`
	NumericScale       *int        `json:"NumericScale,omitempty"`
	DatetimePrecision  *int        `json:"DatetimePrecision,omitempty"`
}

// GetFullDataType returns the data type with size information
func (c *Column) GetFullDataType() string {
	dataType := c.DataType

	// Add character length for string types
	if c.CharacterMaxLength != nil && *c.CharacterMaxLength > 0 {
		switch dataType {
		case "character varying", "varchar", "char", "character", "text":
			if *c.CharacterMaxLength < 2147483647 { // Not MAX length
				return fmt.Sprintf("%s(%d)", dataType, *c.CharacterMaxLength)
			}
		}
	}

	// Add precision and scale for numeric types
	if c.NumericPrecision != nil && *c.NumericPrecision > 0 {
		switch dataType {
		case "numeric", "decimal", "money":
			if c.NumericScale != nil && *c.NumericScale > 0 {
				return fmt.Sprintf("%s(%d,%d)", dataType, *c.NumericPrecision, *c.NumericScale)
			}
			return fmt.Sprintf("%s(%d)", dataType, *c.NumericPrecision)
		}
	}

	// Add precision for datetime types
	if c.DatetimePrecision != nil && *c.DatetimePrecision > 0 {
		switch dataType {
		case "timestamp", "time", "interval":
			return fmt.Sprintf("%s(%d)", dataType, *c.DatetimePrecision)
		}
	}

	return dataType
}

// ForeignKey represents a foreign key constraint
type ForeignKey struct {
	ConstraintName   string `json:"ConstraintName"`
	TableName        string `json:"TableName"`
	ColumnName       string `json:"ColumnName"`
	ReferencedTable  string `json:"ReferencedTable"`
	ReferencedColumn string `json:"ReferencedColumn"`
	UpdateRule       string `json:"UpdateRule"`
	DeleteRule       string `json:"DeleteRule"`
}

// Table represents a database table from the schema
type Table struct {
	TableName   string       `json:"TableName"`
	Columns     []Column     `json:"Columns"`
	ForeignKeys []ForeignKey `json:"ForeignKeys"`
}

// Schema represents the complete database schema
type Schema []Table

// ValidationIssue represents an issue found during validation
type ValidationIssue struct {
	Type       string                 `json:"type" yaml:"type"`
	Severity   string                 `json:"severity" yaml:"severity"`
	Table      string                 `json:"table" yaml:"table"`
	Column     string                 `json:"column,omitempty" yaml:"column,omitempty"`
	Message    string                 `json:"message" yaml:"message"`
	PrimaryKey string                 `json:"primary_key,omitempty" yaml:"primary_key,omitempty"`
	Identifier string                 `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty" yaml:"details,omitempty"`
}

// ValidationReport represents a collection of validation issues
type ValidationReport struct {
	ConnectionName string            `json:"connection_name" yaml:"connection_name"`
	Timestamp      string            `json:"timestamp" yaml:"timestamp"`
	Issues         []ValidationIssue `json:"issues" yaml:"issues"`
	Summary        ReportSummary     `json:"summary" yaml:"summary"`
}

// ReportSummary provides statistics about validation results
type ReportSummary struct {
	TotalIssues   int            `json:"total_issues" yaml:"total_issues"`
	ErrorCount    int            `json:"error_count" yaml:"error_count"`
	WarningCount  int            `json:"warning_count" yaml:"warning_count"`
	TablesCovered int            `json:"tables_covered" yaml:"tables_covered"`
	IssuesByType  map[string]int `json:"issues_by_type" yaml:"issues_by_type"`
}

// SchemaInfo represents schema information for display
type SchemaInfo struct {
	SchemaFile       string         `json:"schema_file" yaml:"schema_file"`
	TotalTables      int            `json:"total_tables" yaml:"total_tables"`
	TotalColumns     int            `json:"total_columns" yaml:"total_columns"`
	TotalForeignKeys int            `json:"total_foreign_keys" yaml:"total_foreign_keys"`
	NotNullColumns   int            `json:"not_null_columns" yaml:"not_null_columns"`
	NullableColumns  int            `json:"nullable_columns" yaml:"nullable_columns"`
	DataTypeCounts   map[string]int `json:"data_type_counts" yaml:"data_type_counts"`
	Tables           []TableSummary `json:"tables" yaml:"tables"`
}

// TableSummary represents a summary of a table for display
type TableSummary struct {
	Name            string `json:"name" yaml:"name"`
	ColumnCount     int    `json:"column_count" yaml:"column_count"`
	ForeignKeyCount int    `json:"foreign_key_count" yaml:"foreign_key_count"`
}

// SchemaComparison represents differences between current and target schema
type SchemaComparison struct {
	MissingTables    []string                   `json:"missing_tables" yaml:"missing_tables"`
	ExtraTables      []string                   `json:"extra_tables" yaml:"extra_tables"`
	TableDifferences map[string]TableDifference `json:"table_differences" yaml:"table_differences"`
}

// TableDifference represents differences in a specific table
type TableDifference struct {
	MissingColumns  []Column              `json:"missing_columns" yaml:"missing_columns"`
	ExtraColumns    []Column              `json:"extra_columns" yaml:"extra_columns"`
	ModifiedColumns map[string]ColumnDiff `json:"modified_columns" yaml:"modified_columns"`
	ForeignKeyDiffs ForeignKeyDifference  `json:"foreign_key_diffs" yaml:"foreign_key_diffs"`
}

// ColumnDiff represents changes in a column definition
type ColumnDiff struct {
	Current Column `json:"current" yaml:"current"`
	Target  Column `json:"target" yaml:"target"`
}

// ForeignKeyDifference represents changes in foreign keys
type ForeignKeyDifference struct {
	Missing []ForeignKey `json:"missing" yaml:"missing"`
	Extra   []ForeignKey `json:"extra" yaml:"extra"`
}

// GetTable returns a table by name from the schema
func (s Schema) GetTable(tableName string) *Table {
	for _, table := range s {
		if table.TableName == tableName {
			return &table
		}
	}
	return nil
}

// GetColumn returns a column by name from the table
func (t *Table) GetColumn(columnName string) *Column {
	for _, column := range t.Columns {
		if column.ColumnName == columnName {
			return &column
		}
	}
	return nil
}

// IsNotNull returns true if the column is marked as NOT NULL
func (c *Column) IsNotNull() bool {
	return c.IsNullable == "NO"
}

// GetPrimaryKeyColumns returns the primary key columns for the table
func (t *Table) GetPrimaryKeyColumns() []Column {
	var pkColumns []Column
	for _, column := range t.Columns {
		// In most schemas, primary keys are typically NOT NULL
		// This is a simplified approach - in a real implementation,
		// you might need to query the database for actual PK constraints
		if column.IsNotNull() && (column.ColumnName == "id" ||
			column.ColumnName == t.TableName+"_id" ||
			column.ColumnName == "uuid" ||
			column.ColumnName == "guid") {
			pkColumns = append(pkColumns, column)
		}
	}
	return pkColumns
}

// FixResult represents the result of a fix operation
type FixResult struct {
	IssuesFound     int    `json:"issues_found"`
	RecordsAffected int    `json:"records_affected"`
	Success         bool   `json:"success"`
	Error           string `json:"error,omitempty"`
	Details         string `json:"details,omitempty"`
}

// FixResults represents results for multiple tables
type FixResults map[string]FixResult
