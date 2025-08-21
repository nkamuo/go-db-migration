package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nkamuo/go-db-migration/internal/models"
)

// LoadSchema loads a schema from a JSON file
func LoadSchema(filePath string) (models.Schema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema models.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	return schema, nil
}

// CompareSchemas compares current database schema with target schema
func CompareSchemas(currentSchema, targetSchema models.Schema) *models.SchemaComparison {
	comparison := &models.SchemaComparison{
		TableDifferences: make(map[string]models.TableDifference),
	}

	// Create maps for easier lookup
	currentTables := make(map[string]*models.Table)
	targetTables := make(map[string]*models.Table)

	for i := range currentSchema {
		currentTables[currentSchema[i].TableName] = &currentSchema[i]
	}

	for i := range targetSchema {
		targetTables[targetSchema[i].TableName] = &targetSchema[i]
	}

	// Find missing and extra tables
	for tableName := range targetTables {
		if _, exists := currentTables[tableName]; !exists {
			comparison.MissingTables = append(comparison.MissingTables, tableName)
		}
	}

	for tableName := range currentTables {
		if _, exists := targetTables[tableName]; !exists {
			comparison.ExtraTables = append(comparison.ExtraTables, tableName)
		}
	}

	// Compare tables that exist in both schemas
	for tableName, targetTable := range targetTables {
		if currentTable, exists := currentTables[tableName]; exists {
			diff := compareTableStructures(currentTable, targetTable)
			if !isTableDifferenceEmpty(diff) {
				comparison.TableDifferences[tableName] = diff
			}
		}
	}

	return comparison
}

// compareTableStructures compares two table structures
func compareTableStructures(currentTable, targetTable *models.Table) models.TableDifference {
	diff := models.TableDifference{
		ModifiedColumns: make(map[string]models.ColumnDiff),
	}

	// Create maps for easier lookup
	currentColumns := make(map[string]*models.Column)
	targetColumns := make(map[string]*models.Column)

	for i := range currentTable.Columns {
		currentColumns[currentTable.Columns[i].ColumnName] = &currentTable.Columns[i]
	}

	for i := range targetTable.Columns {
		targetColumns[targetTable.Columns[i].ColumnName] = &targetTable.Columns[i]
	}

	// Find missing and extra columns
	for columnName, targetColumn := range targetColumns {
		if _, exists := currentColumns[columnName]; !exists {
			diff.MissingColumns = append(diff.MissingColumns, *targetColumn)
		}
	}

	for columnName, currentColumn := range currentColumns {
		if _, exists := targetColumns[columnName]; !exists {
			diff.ExtraColumns = append(diff.ExtraColumns, *currentColumn)
		}
	}

	// Compare columns that exist in both
	for columnName, targetColumn := range targetColumns {
		if currentColumn, exists := currentColumns[columnName]; exists {
			if !areColumnsEqual(currentColumn, targetColumn) {
				diff.ModifiedColumns[columnName] = models.ColumnDiff{
					Current: *currentColumn,
					Target:  *targetColumn,
				}
			}
		}
	}

	// Compare foreign keys
	diff.ForeignKeyDiffs = compareForeignKeys(currentTable.ForeignKeys, targetTable.ForeignKeys)

	return diff
}

// areColumnsEqual compares two columns for equality
func areColumnsEqual(current, target *models.Column) bool {
	return current.DataType == target.DataType &&
		current.IsNullable == target.IsNullable &&
		fmt.Sprintf("%v", current.DefaultValue) == fmt.Sprintf("%v", target.DefaultValue)
}

// compareForeignKeys compares foreign key constraints
func compareForeignKeys(currentFKs, targetFKs []models.ForeignKey) models.ForeignKeyDifference {
	diff := models.ForeignKeyDifference{}

	// Create maps for easier lookup
	currentFKMap := make(map[string]*models.ForeignKey)
	targetFKMap := make(map[string]*models.ForeignKey)

	for i := range currentFKs {
		key := fmt.Sprintf("%s.%s->%s.%s",
			currentFKs[i].TableName, currentFKs[i].ColumnName,
			currentFKs[i].ReferencedTable, currentFKs[i].ReferencedColumn)
		currentFKMap[key] = &currentFKs[i]
	}

	for i := range targetFKs {
		key := fmt.Sprintf("%s.%s->%s.%s",
			targetFKs[i].TableName, targetFKs[i].ColumnName,
			targetFKs[i].ReferencedTable, targetFKs[i].ReferencedColumn)
		targetFKMap[key] = &targetFKs[i]
	}

	// Find missing and extra foreign keys
	for key, targetFK := range targetFKMap {
		if _, exists := currentFKMap[key]; !exists {
			diff.Missing = append(diff.Missing, *targetFK)
		}
	}

	for key, currentFK := range currentFKMap {
		if _, exists := targetFKMap[key]; !exists {
			diff.Extra = append(diff.Extra, *currentFK)
		}
	}

	return diff
}

// isTableDifferenceEmpty checks if a table difference is empty
func isTableDifferenceEmpty(diff models.TableDifference) bool {
	return len(diff.MissingColumns) == 0 &&
		len(diff.ExtraColumns) == 0 &&
		len(diff.ModifiedColumns) == 0 &&
		len(diff.ForeignKeyDiffs.Missing) == 0 &&
		len(diff.ForeignKeyDiffs.Extra) == 0
}

// ValidateSchema performs basic validation on a schema
func ValidateSchema(schema models.Schema) []models.ValidationIssue {
	var issues []models.ValidationIssue

	tableNames := make(map[string]bool)

	for _, table := range schema {
		// Check for duplicate table names
		if tableNames[table.TableName] {
			issues = append(issues, models.ValidationIssue{
				Type:     "duplicate_table",
				Severity: "error",
				Table:    table.TableName,
				Message:  fmt.Sprintf("Duplicate table name: %s", table.TableName),
			})
		}
		tableNames[table.TableName] = true

		// Validate columns
		columnNames := make(map[string]bool)
		for _, column := range table.Columns {
			// Check for duplicate column names
			if columnNames[column.ColumnName] {
				issues = append(issues, models.ValidationIssue{
					Type:     "duplicate_column",
					Severity: "error",
					Table:    table.TableName,
					Column:   column.ColumnName,
					Message:  fmt.Sprintf("Duplicate column name: %s in table %s", column.ColumnName, table.TableName),
				})
			}
			columnNames[column.ColumnName] = true

			// Check for missing column name
			if column.ColumnName == "" {
				issues = append(issues, models.ValidationIssue{
					Type:     "invalid_column",
					Severity: "error",
					Table:    table.TableName,
					Message:  "Column with empty name found",
				})
			}

			// Check for missing data type
			if column.DataType == "" {
				issues = append(issues, models.ValidationIssue{
					Type:     "invalid_column",
					Severity: "error",
					Table:    table.TableName,
					Column:   column.ColumnName,
					Message:  fmt.Sprintf("Column %s has no data type", column.ColumnName),
				})
			}
		}

		// Validate foreign keys reference valid tables and columns
		for _, fk := range table.ForeignKeys {
			// Check if referenced table exists in schema
			referencedTable := schema.GetTable(fk.ReferencedTable)
			if referencedTable == nil {
				issues = append(issues, models.ValidationIssue{
					Type:     "invalid_foreign_key",
					Severity: "warning",
					Table:    table.TableName,
					Column:   fk.ColumnName,
					Message:  fmt.Sprintf("Foreign key references non-existent table: %s", fk.ReferencedTable),
					Details: map[string]interface{}{
						"constraint_name":   fk.ConstraintName,
						"referenced_table":  fk.ReferencedTable,
						"referenced_column": fk.ReferencedColumn,
					},
				})
			} else {
				// Check if referenced column exists
				referencedColumn := referencedTable.GetColumn(fk.ReferencedColumn)
				if referencedColumn == nil {
					issues = append(issues, models.ValidationIssue{
						Type:     "invalid_foreign_key",
						Severity: "warning",
						Table:    table.TableName,
						Column:   fk.ColumnName,
						Message:  fmt.Sprintf("Foreign key references non-existent column: %s.%s", fk.ReferencedTable, fk.ReferencedColumn),
						Details: map[string]interface{}{
							"constraint_name":   fk.ConstraintName,
							"referenced_table":  fk.ReferencedTable,
							"referenced_column": fk.ReferencedColumn,
						},
					})
				}
			}

			// Check if source column exists
			sourceColumn := table.GetColumn(fk.ColumnName)
			if sourceColumn == nil {
				issues = append(issues, models.ValidationIssue{
					Type:     "invalid_foreign_key",
					Severity: "error",
					Table:    table.TableName,
					Column:   fk.ColumnName,
					Message:  fmt.Sprintf("Foreign key references non-existent source column: %s", fk.ColumnName),
					Details: map[string]interface{}{
						"constraint_name": fk.ConstraintName,
					},
				})
			}
		}
	}

	return issues
}
