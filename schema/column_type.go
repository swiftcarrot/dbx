package schema

// ColumnType represents a SQL column type
// This interface ensures only valid SQL column types can be used
// Each database dialect may provide its own implementations
type ColumnType interface {
	SQL() string
}

// TextType represents a text column type
type TextType struct{}

func (t *TextType) SQL() string {
	return "text"
}

// IntegerType represents an integer column type
type IntegerType struct{}

func (t *IntegerType) SQL() string {
	return "integer"
}

// BigIntType represents a bigint column type
type BigIntType struct{}

func (t *BigIntType) SQL() string {
	return "bigint"
}

// SmallIntType represents a smallint column type
type SmallIntType struct{}

func (t *SmallIntType) SQL() string {
	return "smallint"
}

// BooleanType represents a boolean column type
type BooleanType struct{}

func (t *BooleanType) SQL() string {
	return "boolean"
}

// FloatType represents a float column type
type FloatType struct{}

func (t *FloatType) SQL() string {
	return "float"
}

// DecimalType represents a decimal column type
type DecimalType struct {
	Precision int
	Scale     int
}

func (t *DecimalType) SQL() string {
	return "decimal"
}

// VarcharType represents a varchar column type
type VarcharType struct {
	Length int
}

func (t *VarcharType) SQL() string {
	return "varchar"
}

// TimestampType represents a timestamp column type
type TimestampType struct {
	WithTimeZone bool
}

func (t *TimestampType) SQL() string {
	if t.WithTimeZone {
		return "timestamp with time zone"
	}
	return "timestamp"
}

// DateType represents a date column type
type DateType struct{}

func (t *DateType) SQL() string {
	return "date"
}

// TimeType represents a time column type
type TimeType struct{}

func (t *TimeType) SQL() string {
	return "time"
}

// UUIDType represents a UUID column type
type UUIDType struct{}

func (t *UUIDType) SQL() string {
	return "uuid"
}

// BlobType represents a blob column type
type BlobType struct{}

func (t *BlobType) SQL() string {
	return "blob"
}
