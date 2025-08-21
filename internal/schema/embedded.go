package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nkamuo/go-db-migration/internal/models"
)

//go:embed schema.json
var embeddedSchema []byte

// LoadSchemaWithFallback loads schema with embedded fallback
func LoadSchemaWithFallback(filePath string) (models.Schema, error) {
	// First try to load from the specified file path
	if filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			return LoadSchema(filePath)
		}
	}

	// Fall back to embedded schema
	return LoadEmbeddedSchema()
}

// LoadEmbeddedSchema loads the embedded schema
func LoadEmbeddedSchema() (models.Schema, error) {
	var schema models.Schema
	if err := json.Unmarshal(embeddedSchema, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse embedded schema JSON: %w", err)
	}
	return schema, nil
}

// GetEmbeddedSchemaPath returns a descriptive path for the embedded schema
func GetEmbeddedSchemaPath() string {
	return "embedded://schema.json"
}
