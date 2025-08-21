package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	"github.com/nkamuo/go-db-migration/internal/models"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatYAML  OutputFormat = "yaml"
	FormatCSV   OutputFormat = "csv"
)

// Formatter handles different output formats
type Formatter struct {
	format OutputFormat
}

// NewFormatter creates a new output formatter
func NewFormatter(format string) *Formatter {
	return &Formatter{format: OutputFormat(format)}
}

// FormatValidationReport formats a validation report in the specified format
func (f *Formatter) FormatValidationReport(report *models.ValidationReport) (string, error) {
	switch f.format {
	case FormatTable:
		return f.formatValidationReportAsTable(report), nil
	case FormatJSON:
		return f.formatValidationReportAsJSON(report)
	case FormatYAML:
		return f.formatValidationReportAsYAML(report)
	case FormatCSV:
		return f.formatValidationReportAsCSV(report)
	default:
		return "", fmt.Errorf("unsupported output format: %s", f.format)
	}
}

// FormatSchemaInfo formats schema information in the specified format
func (f *Formatter) FormatSchemaInfo(info *models.SchemaInfo) (string, error) {
	switch f.format {
	case FormatTable:
		return f.formatSchemaInfoAsTable(info), nil
	case FormatJSON:
		return f.formatSchemaInfoAsJSON(info)
	case FormatYAML:
		return f.formatSchemaInfoAsYAML(info)
	default:
		return "", fmt.Errorf("unsupported output format for schema info: %s", f.format)
	}
}

// FormatSchemaComparison formats a schema comparison in the specified format
func (f *Formatter) FormatSchemaComparison(comparison *models.SchemaComparison) (string, error) {
	switch f.format {
	case FormatTable:
		return f.formatSchemaComparisonAsTable(comparison), nil
	case FormatJSON:
		return f.formatSchemaComparisonAsJSON(comparison)
	case FormatYAML:
		return f.formatSchemaComparisonAsYAML(comparison)
	default:
		return "", fmt.Errorf("unsupported output format for schema comparison: %s", f.format)
	}
}

// formatValidationReportAsTable formats the validation report as a table
func (f *Formatter) formatValidationReportAsTable(report *models.ValidationReport) string {
	if len(report.Issues) == 0 {
		return "✅ No validation issues found!\n"
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetTitle("Database Validation Report")
	t.AppendHeader(table.Row{"Severity", "Type", "Table", "Column", "Message", "Identifier"})

	for _, issue := range report.Issues {
		severity := issue.Severity
		if severity == "error" {
			severity = text.FgHiRed.Sprint("ERROR")
		} else if severity == "warning" {
			severity = text.FgHiYellow.Sprint("WARNING")
		}

		t.AppendRow(table.Row{
			severity,
			issue.Type,
			issue.Table,
			issue.Column,
			issue.Message,
			issue.Identifier,
		})
	}

	t.AppendSeparator()
	t.AppendFooter(table.Row{
		"Summary",
		fmt.Sprintf("Total: %d", report.Summary.TotalIssues),
		fmt.Sprintf("Errors: %d", report.Summary.ErrorCount),
		fmt.Sprintf("Warnings: %d", report.Summary.WarningCount),
		fmt.Sprintf("Tables: %d", report.Summary.TablesCovered),
		fmt.Sprintf("Connection: %s", report.ConnectionName),
	})

	t.SetStyle(table.StyleColoredBright)
	return t.Render()
}

// formatValidationReportAsJSON formats the validation report as JSON
func (f *Formatter) formatValidationReportAsJSON(report *models.ValidationReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal validation report to JSON: %w", err)
	}
	return string(data), nil
}

// formatValidationReportAsYAML formats the validation report as YAML
func (f *Formatter) formatValidationReportAsYAML(report *models.ValidationReport) (string, error) {
	data, err := yaml.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("failed to marshal validation report to YAML: %w", err)
	}
	return string(data), nil
}

// formatValidationReportAsCSV formats the validation report as CSV
func (f *Formatter) formatValidationReportAsCSV(report *models.ValidationReport) (string, error) {
	records := [][]string{
		{"Severity", "Type", "Table", "Column", "Message", "Identifier", "PrimaryKey"},
	}

	for _, issue := range report.Issues {
		records = append(records, []string{
			issue.Severity,
			issue.Type,
			issue.Table,
			issue.Column,
			issue.Message,
			issue.Identifier,
			issue.PrimaryKey,
		})
	}

	// Convert to CSV string
	var result string
	for _, record := range records {
		for i, field := range record {
			if i > 0 {
				result += ","
			}
			result += fmt.Sprintf(`"%s"`, field)
		}
		result += "\n"
	}

	return result, nil
}

// formatSchemaComparisonAsTable formats the schema comparison as a table
func (f *Formatter) formatSchemaComparisonAsTable(comparison *models.SchemaComparison) string {
	var output string

	// Missing tables
	if len(comparison.MissingTables) > 0 {
		t := table.NewWriter()
		t.SetTitle("Missing Tables")
		t.AppendHeader(table.Row{"Table Name"})
		for _, tableName := range comparison.MissingTables {
			t.AppendRow(table.Row{text.FgRed.Sprint(tableName)})
		}
		t.SetStyle(table.StyleColoredBright)
		output += t.Render() + "\n\n"
	}

	// Extra tables
	if len(comparison.ExtraTables) > 0 {
		t := table.NewWriter()
		t.SetTitle("Extra Tables")
		t.AppendHeader(table.Row{"Table Name"})
		for _, tableName := range comparison.ExtraTables {
			t.AppendRow(table.Row{text.FgYellow.Sprint(tableName)})
		}
		t.SetStyle(table.StyleColoredBright)
		output += t.Render() + "\n\n"
	}

	// Table differences
	for tableName, diff := range comparison.TableDifferences {
		if len(diff.MissingColumns) > 0 || len(diff.ExtraColumns) > 0 || len(diff.ModifiedColumns) > 0 {
			t := table.NewWriter()
			t.SetTitle(fmt.Sprintf("Table: %s", tableName))
			t.AppendHeader(table.Row{"Change Type", "Column", "Details"})

			for _, col := range diff.MissingColumns {
				t.AppendRow(table.Row{
					text.FgRed.Sprint("MISSING"),
					col.ColumnName,
					fmt.Sprintf("%s, %s", col.DataType, col.IsNullable),
				})
			}

			for _, col := range diff.ExtraColumns {
				t.AppendRow(table.Row{
					text.FgYellow.Sprint("EXTRA"),
					col.ColumnName,
					fmt.Sprintf("%s, %s", col.DataType, col.IsNullable),
				})
			}

			for colName, colDiff := range diff.ModifiedColumns {
				t.AppendRow(table.Row{
					text.FgBlue.Sprint("MODIFIED"),
					colName,
					fmt.Sprintf("Current: %s (%s) → Target: %s (%s)",
						colDiff.Current.DataType, colDiff.Current.IsNullable,
						colDiff.Target.DataType, colDiff.Target.IsNullable),
				})
			}

			t.SetStyle(table.StyleColoredBright)
			output += t.Render() + "\n\n"
		}
	}

	if output == "" {
		output = "✅ No schema differences found!\n"
	}

	return output
}

// formatSchemaComparisonAsJSON formats the schema comparison as JSON
func (f *Formatter) formatSchemaComparisonAsJSON(comparison *models.SchemaComparison) (string, error) {
	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema comparison to JSON: %w", err)
	}
	return string(data), nil
}

// formatSchemaComparisonAsYAML formats the schema comparison as YAML
func (f *Formatter) formatSchemaComparisonAsYAML(comparison *models.SchemaComparison) (string, error) {
	data, err := yaml.Marshal(comparison)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema comparison to YAML: %w", err)
	}
	return string(data), nil
}

// formatSchemaInfoAsTable formats the schema info as a table
func (f *Formatter) formatSchemaInfoAsTable(info *models.SchemaInfo) string {
	var output string

	// Summary information table
	summaryTable := table.NewWriter()
	summaryTable.SetTitle("📊 Schema Summary")
	summaryTable.AppendHeader(table.Row{"Metric", "Value"})
	
	summaryTable.AppendRow(table.Row{"📁 Schema File", info.SchemaFile})
	summaryTable.AppendRow(table.Row{"🗂️  Total Tables", info.TotalTables})
	summaryTable.AppendRow(table.Row{"📋 Total Columns", info.TotalColumns})
	summaryTable.AppendRow(table.Row{"🔗 Foreign Keys", info.TotalForeignKeys})
	summaryTable.AppendRow(table.Row{"🚫 NOT NULL Columns", info.NotNullColumns})
	summaryTable.AppendRow(table.Row{"✅ Nullable Columns", info.NullableColumns})
	
	summaryTable.SetStyle(table.StyleColoredBright)
	output += summaryTable.Render() + "\n\n"

	// Data types table
	if len(info.DataTypeCounts) > 0 {
		dataTypesTable := table.NewWriter()
		dataTypesTable.SetTitle("📊 Data Types Distribution")
		dataTypesTable.AppendHeader(table.Row{"Data Type", "Count"})
		
		for dataType, count := range info.DataTypeCounts {
			dataTypesTable.AppendRow(table.Row{dataType, count})
		}
		
		dataTypesTable.SetStyle(table.StyleColoredBright)
		output += dataTypesTable.Render() + "\n\n"
	}

	// Tables detail
	if len(info.Tables) > 0 {
		tablesTable := table.NewWriter()
		tablesTable.SetTitle("📋 Tables Detail")
		tablesTable.AppendHeader(table.Row{"Table Name", "Columns", "Foreign Keys"})
		
		for _, tableSummary := range info.Tables {
			tablesTable.AppendRow(table.Row{
				tableSummary.Name,
				tableSummary.ColumnCount,
				tableSummary.ForeignKeyCount,
			})
		}
		
		tablesTable.SetStyle(table.StyleColoredBright)
		output += tablesTable.Render()
	}

	return output
}

// formatSchemaInfoAsJSON formats the schema info as JSON
func (f *Formatter) formatSchemaInfoAsJSON(info *models.SchemaInfo) (string, error) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema info to JSON: %w", err)
	}
	return string(data), nil
}

// formatSchemaInfoAsYAML formats the schema info as YAML
func (f *Formatter) formatSchemaInfoAsYAML(info *models.SchemaInfo) (string, error) {
	data, err := yaml.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema info to YAML: %w", err)
	}
	return string(data), nil
}

// WriteToFile writes content to a file
func WriteToFile(content, filename string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// CreateValidationReport creates a validation report with summary
func CreateValidationReport(connectionName string, issues []models.ValidationIssue) *models.ValidationReport {
	summary := models.ReportSummary{
		TotalIssues:  len(issues),
		IssuesByType: make(map[string]int),
	}

	tables := make(map[string]bool)
	for _, issue := range issues {
		if issue.Severity == "error" {
			summary.ErrorCount++
		} else if issue.Severity == "warning" {
			summary.WarningCount++
		}

		summary.IssuesByType[issue.Type]++
		if issue.Table != "" {
			tables[issue.Table] = true
		}
	}

	summary.TablesCovered = len(tables)

	return &models.ValidationReport{
		ConnectionName: connectionName,
		Timestamp:      time.Now().Format(time.RFC3339),
		Issues:         issues,
		Summary:        summary,
	}
}

// SaveReportToFile saves a report to a file with the specified format
func SaveReportToFile(report *models.ValidationReport, filename string, format OutputFormat) error {
	formatter := NewFormatter(string(format))
	content, err := formatter.FormatValidationReport(report)
	if err != nil {
		return err
	}

	return WriteToFile(content, filename)
}

// SaveComparisonToFile saves a schema comparison to a file with the specified format
func SaveComparisonToFile(comparison *models.SchemaComparison, filename string, format OutputFormat) error {
	formatter := NewFormatter(string(format))
	content, err := formatter.FormatSchemaComparison(comparison)
	if err != nil {
		return err
	}

	return WriteToFile(content, filename)
}

// CreateSchemaInfo creates a SchemaInfo from a schema
func CreateSchemaInfo(schemaFile string, schema models.Schema) *models.SchemaInfo {
	info := &models.SchemaInfo{
		SchemaFile:     schemaFile,
		TotalTables:    len(schema),
		DataTypeCounts: make(map[string]int),
		Tables:         make([]models.TableSummary, 0, len(schema)),
	}

	for _, table := range schema {
		// Count columns and foreign keys
		info.TotalColumns += len(table.Columns)
		info.TotalForeignKeys += len(table.ForeignKeys)

		// Count nullable vs not null columns
		for _, column := range table.Columns {
			if column.IsNotNull() {
				info.NotNullColumns++
			} else {
				info.NullableColumns++
			}

			// Count data types
			info.DataTypeCounts[column.DataType]++
		}

		// Add table summary
		info.Tables = append(info.Tables, models.TableSummary{
			Name:            table.TableName,
			ColumnCount:     len(table.Columns),
			ForeignKeyCount: len(table.ForeignKeys),
		})
	}

	return info
}
