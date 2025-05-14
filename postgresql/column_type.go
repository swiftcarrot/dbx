package postgresql

import "github.com/swiftcarrot/dbx/schema"

// JSONType represents a JSON column type in PostgreSQL
type JSONType struct{}

func (t *JSONType) SQL() string {
	return "json"
}

// JSONBType represents a JSONB column type in PostgreSQL
type JSONBType struct{}

func (t *JSONBType) SQL() string {
	return "jsonb"
}

// ArrayType represents an array column type in PostgreSQL
type ArrayType struct {
	ElementType schema.ColumnType
}

func (t *ArrayType) SQL() string {
	return t.ElementType.SQL() + "[]"
}

// IntervalType represents an INTERVAL column type in PostgreSQL
type IntervalType struct{}

func (t *IntervalType) SQL() string {
	return "interval"
}

// CIDRType represents a CIDR column type in PostgreSQL
type CIDRType struct{}

func (t *CIDRType) SQL() string {
	return "cidr"
}

// INETType represents an INET column type in PostgreSQL
type INETType struct{}

func (t *INETType) SQL() string {
	return "inet"
}

// MACAddrType represents a MACADDR column type in PostgreSQL
type MACAddrType struct{}

func (t *MACAddrType) SQL() string {
	return "macaddr"
}

// SerialType represents a SERIAL column type in PostgreSQL
type SerialType struct{}

func (t *SerialType) SQL() string {
	return "serial"
}

// BigSerialType represents a BIGSERIAL column type in PostgreSQL
type BigSerialType struct{}

func (t *BigSerialType) SQL() string {
	return "bigserial"
}
