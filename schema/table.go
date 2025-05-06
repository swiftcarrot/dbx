package schema

// Table represents a database table
type Table struct {
	Schema      string
	Name        string
	Columns     []*Column
	Indexes     []*Index
	PrimaryKey  *PrimaryKey
	ForeignKeys []*ForeignKey
}

// Column adds a column to a table
func (t *Table) Column(name string, columnType string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: columnType,
	}

	for _, option := range options {
		option(col)
	}

	t.Columns = append(t.Columns, col)
	return col
}

// Index adds an index to a table
func (t *Table) Index(name string, columns []string, options ...IndexOption) *Index {
	idx := &Index{
		Name:    name,
		Columns: columns,
	}

	for _, option := range options {
		option(idx)
	}

	t.Indexes = append(t.Indexes, idx)
	return idx
}

// SetPrimaryKey sets the primary key for a table
func (t *Table) SetPrimaryKey(name string, columns []string) *PrimaryKey {
	t.PrimaryKey = &PrimaryKey{
		Name:    name,
		Columns: columns,
	}
	return t.PrimaryKey
}

// ForeignKey adds a foreign key to a table
func (t *Table) ForeignKey(name string, columns []string, refTable string, refColumns []string, options ...ForeignKeyOption) *ForeignKey {
	fk := &ForeignKey{
		Name:       name,
		Columns:    columns,
		RefTable:   refTable,
		RefColumns: refColumns,
	}

	for _, option := range options {
		option(fk)
	}

	t.ForeignKeys = append(t.ForeignKeys, fk)
	return fk
}

// Column represents a table column
type Column struct {
	Name      string
	Type      string
	Nullable  bool
	Default   string
	Precision int
	Scale     int
	Comment   string

	Identity      string
	PrimaryKey    bool
	AutoIncrement bool
	Length        int
}

// ColumnOption is a function type for column options
type ColumnOption func(*Column)

// Nullable makes a column nullable
func Nullable(c *Column) {
	c.Nullable = true
}

// NotNull makes a column not nullable
func NotNull(c *Column) {
	c.Nullable = false
}

// Default sets a default value for a column
func Default(value string) ColumnOption {
	return func(c *Column) {
		c.Default = value
	}
}

// Comment sets a comment for a column
func Comment(comment string) ColumnOption {
	return func(c *Column) {
		c.Comment = comment
	}
}

// Index represents a table index
type Index struct {
	Name    string
	Columns []string
	Unique  bool
}

// IndexOption is a function type for index options
type IndexOption func(*Index)

// Unique makes an index unique
func Unique(i *Index) {
	i.Unique = true
}

// PrimaryKey represents a table's primary key
type PrimaryKey struct {
	Name    string
	Columns []string
}
