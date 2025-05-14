package sqlite

// NumericType represents a NUMERIC column type in SQLite
type NumericType struct{}

func (t *NumericType) SQL() string {
	return "numeric"
}

// RealType represents a REAL column type in SQLite
type RealType struct{}

func (t *RealType) SQL() string {
	return "real"
}

// IntegerType represents an INTEGER column type in SQLite (uppercase)
type IntegerType struct{}

func (t *IntegerType) SQL() string {
	return "INTEGER"
}

// TextType represents a TEXT column type in SQLite (uppercase)
type TextType struct{}

func (t *TextType) SQL() string {
	return "TEXT"
}

// TimestampType represents a TIMESTAMP column type in SQLite (uppercase)
type TimestampType struct{}

func (t *TimestampType) SQL() string {
	return "TIMESTAMP"
}
