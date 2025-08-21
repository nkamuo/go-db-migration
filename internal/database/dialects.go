package database

import (
	"fmt"
	"github.com/nkamuo/go-db-migration/internal/config"
	"github.com/nkamuo/go-db-migration/internal/models"
)

// PostgreSQLDialect implements PostgreSQL-specific queries
type PostgreSQLDialect struct{}

func (d *PostgreSQLDialect) GetDriverName() string {
	return "postgres"
}

func (d *PostgreSQLDialect) GetIdentifierQuote() string {
	return `"`
}

func (d *PostgreSQLDialect) BuildConnectionString(cfg *config.DBConfig) string {
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, sslmode)
}

func (d *PostgreSQLDialect) GetTablesQuery() string {
	return `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`
}

func (d *PostgreSQLDialect) GetColumnsQuery() string {
	return `
		SELECT 
			column_name,
			data_type,
			column_default,
			is_nullable
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		  AND table_name = $1
		ORDER BY ordinal_position`
}

func (d *PostgreSQLDialect) GetForeignKeysQuery() string {
	return `
		SELECT 
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			rc.update_rule,
			rc.delete_rule
		FROM information_schema.table_constraints AS tc 
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		  AND tc.table_name = $1`
}

func (d *PostgreSQLDialect) GetTableRowCountQuery(tableName string) string {
	return fmt.Sprintf(`SELECT COUNT(*) FROM "%s"`, tableName)
}

func (d *PostgreSQLDialect) GetNullViolationsQuery(tableName, columnName, identifierCol string) string {
	return fmt.Sprintf(`
		SELECT "%s"
		FROM "%s"
		WHERE "%s" IS NULL
		LIMIT 1000`, identifierCol, tableName, columnName)
}

func (d *PostgreSQLDialect) GetForeignKeyViolationsQuery(fk models.ForeignKey, identifierCol string) string {
	return fmt.Sprintf(`
		SELECT "%s", "%s"
		FROM "%s" t1
		WHERE "%s" IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM "%s" t2 
			WHERE t2."%s" = t1."%s"
		  )
		LIMIT 1000`,
		fk.ColumnName, identifierCol, fk.TableName, fk.ColumnName,
		fk.ReferencedTable, fk.ReferencedColumn, fk.ColumnName)
}

// MySQLDialect implements MySQL-specific queries
type MySQLDialect struct{}

func (d *MySQLDialect) GetDriverName() string {
	return "mysql"
}

func (d *MySQLDialect) GetIdentifierQuote() string {
	return "`"
}

func (d *MySQLDialect) BuildConnectionString(cfg *config.DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
}

func (d *MySQLDialect) GetTablesQuery() string {
	return `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`
}

func (d *MySQLDialect) GetColumnsQuery() string {
	return `
		SELECT 
			column_name,
			data_type,
			column_default,
			is_nullable
		FROM information_schema.columns 
		WHERE table_schema = DATABASE()
		  AND table_name = ?
		ORDER BY ordinal_position`
}

func (d *MySQLDialect) GetForeignKeysQuery() string {
	return `
		SELECT 
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			kcu.referenced_table_name AS foreign_table_name,
			kcu.referenced_column_name AS foreign_column_name,
			rc.update_rule,
			rc.delete_rule
		FROM information_schema.table_constraints AS tc 
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.referential_constraints AS rc
			ON tc.constraint_name = rc.constraint_name
			AND tc.table_schema = rc.constraint_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' 
		  AND tc.table_schema = DATABASE()
		  AND tc.table_name = ?`
}

func (d *MySQLDialect) GetTableRowCountQuery(tableName string) string {
	return fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
}

func (d *MySQLDialect) GetNullViolationsQuery(tableName, columnName, identifierCol string) string {
	return fmt.Sprintf(`
		SELECT `+"`%s`"+`
		FROM `+"`%s`"+`
		WHERE `+"`%s`"+` IS NULL
		LIMIT 1000`, identifierCol, tableName, columnName)
}

func (d *MySQLDialect) GetForeignKeyViolationsQuery(fk models.ForeignKey, identifierCol string) string {
	return fmt.Sprintf(`
		SELECT `+"`%s`, `%s`"+`
		FROM `+"`%s`"+` t1
		WHERE `+"`%s`"+` IS NOT NULL
		  AND NOT EXISTS (
			SELECT 1 FROM `+"`%s`"+` t2 
			WHERE t2.`+"`%s`"+` = t1.`+"`%s`"+`
		  )
		LIMIT 1000`,
		fk.ColumnName, identifierCol, fk.TableName, fk.ColumnName,
		fk.ReferencedTable, fk.ReferencedColumn, fk.ColumnName)
}
