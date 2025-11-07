package db

import (
	"database/sql"
	"os"
	"regexp"

	"github.com/novapix/rssbot/logger"

	_ "github.com/lib/pq"
)

func InitDB(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		logger.Error.Fatalf("DB open error: %v", err)
	}

	if err := db.Ping(); err != nil {
		logger.Error.Fatalf("DB ping error: %v", err)
	}

	logger.Info.Println("Connected to PostgreSQL")

	schemaFile := "db/schema.sql"
	sqlBytes, err := os.ReadFile(schemaFile)
	if err != nil {
		logger.Error.Fatalf("Failed to read schema.sql: %v", err)
	}
	schemaSQL := string(sqlBytes)

	// Extract required tables dynamically
	requiredTables := extractTablesFromSQL(schemaSQL)
	logger.Info.Printf("Required tables: %v", requiredTables)

	// Verify or create schema
	if err := verifyOrCreateSchema(db, requiredTables, schemaSQL); err != nil {
		logger.Error.Fatalf("Schema verification failed: %v", err)
	}

	return db
}

func extractTablesFromSQL(schemaSQL string) []string {
	re := regexp.MustCompile(`(?i)CREATE TABLE IF NOT EXISTS\s+(\w+)`)
	matches := re.FindAllStringSubmatch(schemaSQL, -1)

	var tables []string
	for _, m := range matches {
		if len(m) > 1 {
			tables = append(tables, m[1])
		}
	}
	return tables
}

func verifyOrCreateSchema(db *sql.DB, tables []string, schemaSQL string) error {
	missing := []string{}
	for _, t := range tables {
		if !tableExists(db, t) {
			missing = append(missing, t)
		}
	}

	if len(missing) == 0 {
		logger.Info.Println("All required tables exist")
		return nil
	}

	logger.Info.Printf("Missing tables: %v. Attempting to create...", missing)

	_, err := db.Exec(schemaSQL)
	if err != nil {
		return err
	}

	for _, t := range tables {
		if !tableExists(db, t) {
			return &SchemaError{Table: t}
		}
	}

	logger.Info.Println("Missing tables created successfully")
	return nil
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	)`
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		logger.Error.Printf("Error checking table %s: %v", tableName, err)
		return false
	}
	return exists
}

type SchemaError struct {
	Table string
}

func (e *SchemaError) Error() string {
	return "Table " + e.Table + " does not exist after schema creation"
}
