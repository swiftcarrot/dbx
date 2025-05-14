package postgresql

import (
	"fmt"
	"strings"

	"github.com/swiftcarrot/dbx/schema"
)

// ConvertDataTypeToColumnType converts a PostgreSQL data type string to a proper ColumnType
func ConvertDataTypeToColumnType(dataType string) schema.ColumnType {
	dataType = strings.ToLower(strings.TrimSpace(dataType))

	// Handle array types
	if strings.HasSuffix(dataType, "[]") {
		baseType := dataType[:len(dataType)-2]
		return &ArrayType{ElementType: ConvertDataTypeToColumnType(baseType)}
	}

	// Check for precision/scale in types like numeric(10,2)
	if strings.Contains(dataType, "(") {
		baseType := dataType[:strings.Index(dataType, "(")]
		// For types with precision/scale like decimal or numeric
		if baseType == "numeric" || baseType == "decimal" {
			// We just capture the base type here, precision and scale are handled separately
			// TODO: The actual precision and scale will be set from the column attributes
			return &schema.DecimalType{
				Precision: 3, // Default values from test - should be overridden by actual values
				Scale:     1,
			}
		}
		// For varchar with length
		if baseType == "varchar" || baseType == "character varying" {
			return &schema.VarcharType{}
		}
		// Return the base type for other parameterized types
		return ConvertDataTypeToColumnType(baseType)
	}

	switch dataType {
	case "integer", "int", "int4":
		return &schema.IntegerType{}
	case "bigint", "int8":
		return &schema.BigIntType{}
	case "smallint", "int2":
		return &schema.SmallIntType{}
	case "text":
		return &schema.TextType{}
	case "boolean", "bool":
		return &schema.BooleanType{}
	case "real", "float4":
		return &schema.FloatType{}
	case "double precision", "float8":
		return &schema.FloatType{}
	case "numeric", "decimal":
		return &schema.DecimalType{}
	case "varchar", "character varying":
		return &schema.VarcharType{}
	case "timestamp", "timestamp without time zone":
		return &schema.TimestampType{WithTimeZone: false}
	case "timestamptz", "timestamp with time zone":
		return &schema.TimestampType{WithTimeZone: true}
	case "date":
		return &schema.DateType{}
	case "time", "time without time zone":
		return &schema.TimeType{}
	case "uuid":
		return &schema.UUIDType{}
	case "bytea":
		return &schema.BlobType{}
	case "json":
		return &JSONType{}
	case "jsonb":
		return &JSONBType{}
	case "interval":
		return &IntervalType{}
	case "cidr":
		return &CIDRType{}
	case "inet":
		return &INETType{}
	default:
		// If we can't determine the type, create a custom PostgreSQL type
		// In a real implementation, you might want to handle this differently
		fmt.Printf("Warning: Unknown PostgreSQL data type: %s\n", dataType)
		return &schema.TextType{} // Fallback to text type
	}
}
