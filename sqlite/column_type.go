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
