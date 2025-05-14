package mysql

// IntType represents an INT column type in MySQL (instead of INTEGER)
type IntType struct{}

func (t *IntType) SQL() string {
	return "int"
}

// JSONType represents a JSON column type in MySQL
type JSONType struct{}

func (t *JSONType) SQL() string {
	return "json"
}

// ENUMType represents an ENUM column type in MySQL
type ENUMType struct {
	Values []string
}

func (t *ENUMType) SQL() string {
	return "enum"
}

// SetType represents a SET column type in MySQL
type SetType struct {
	Values []string
}

func (t *SetType) SQL() string {
	return "set"
}

// TinyIntType represents a TINYINT column type in MySQL
type TinyIntType struct{}

func (t *TinyIntType) SQL() string {
	return "tinyint"
}

// MediumIntType represents a MEDIUMINT column type in MySQL
type MediumIntType struct{}

func (t *MediumIntType) SQL() string {
	return "mediumint"
}

// TinyTextType represents a TINYTEXT column type in MySQL
type TinyTextType struct{}

func (t *TinyTextType) SQL() string {
	return "tinytext"
}

// MediumTextType represents a MEDIUMTEXT column type in MySQL
type MediumTextType struct{}

func (t *MediumTextType) SQL() string {
	return "mediumtext"
}

// LongTextType represents a LONGTEXT column type in MySQL
type LongTextType struct{}

func (t *LongTextType) SQL() string {
	return "longtext"
}
