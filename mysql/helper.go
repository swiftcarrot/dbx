package mysql

import (
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// ConvertDataTypeToColumnType converts a MySQL data type string to a proper ColumnType
func ConvertDataTypeToColumnType(dataType string) schema.ColumnType {
	dataType = strings.ToLower(strings.TrimSpace(dataType))

	// Check for precision/scale in types like numeric(10,2)
	if strings.Contains(dataType, "(") {
		baseType := dataType[:strings.Index(dataType, "(")]
		// For types with precision/scale like decimal or numeric
		if baseType == "numeric" || baseType == "decimal" {
			// We just capture the base type here, precision and scale are handled separately
			return &schema.DecimalType{}
		}
		// For varchar with length
		if baseType == "varchar" || baseType == "char" {
			return &schema.VarcharType{}
		}
		// Return the base type for other parameterized types
		return ConvertDataTypeToColumnType(baseType)
	}

	switch dataType {
	case "int", "integer":
		return &schema.IntegerType{}
	case "bigint":
		return &schema.BigIntType{}
	case "smallint":
		return &schema.SmallIntType{}
	case "tinyint":
		return &TinyIntType{}
	case "mediumint":
		return &MediumIntType{}
	case "text", "longtext":
		return &schema.TextType{}
	case "tinytext":
		return &TinyTextType{}
	case "mediumtext":
		return &MediumTextType{}
	case "boolean", "bool", "tinyint(1)":
		return &schema.BooleanType{}
	case "float":
		return &schema.FloatType{}
	case "double":
		return &schema.FloatType{}
	case "numeric", "decimal":
		return &schema.DecimalType{}
	case "varchar", "char":
		return &schema.VarcharType{}
	case "timestamp":
		return &schema.TimestampType{WithTimeZone: false}
	case "datetime":
		return &schema.TimestampType{WithTimeZone: false}
	case "date":
		return &schema.DateType{}
	case "time":
		return &schema.TimeType{}
	case "blob", "longblob":
		return &schema.BlobType{}
	case "json":
		return &JSONType{}
	default:
		// If it's an ENUM or SET type
		if strings.HasPrefix(dataType, "enum") {
			return &ENUMType{}
		}
		if strings.HasPrefix(dataType, "set") {
			return &SetType{}
		}

		// If we can't determine the type, create a fallback
		fmt.Printf("Warning: Unknown MySQL data type: %s\n", dataType)
		return &schema.TextType{} // Fallback to text type
	}
}
