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
func (t *Table) Column(name string, columnType ColumnType, options ...ColumnOption) *Column {
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

// String adds a varchar column (default length 255)
func (t *Table) String(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &VarcharType{Length: 255},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Text adds a text column
func (t *Table) Text(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &TextType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Integer adds an integer column
func (t *Table) Integer(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &IntegerType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// BigInt adds a bigint column
func (t *Table) BigInt(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &BigIntType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Float adds a float column
func (t *Table) Float(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &FloatType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Decimal adds a decimal column (default precision/scale 0)
func (t *Table) Decimal(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &DecimalType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// DateTime adds a timestamp column
func (t *Table) DateTime(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &TimestampType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Time adds a time column
func (t *Table) Time(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &TimeType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Date adds a date column
func (t *Table) Date(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &DateType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Binary adds a blob column
func (t *Table) Binary(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &BlobType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Boolean adds a boolean column
func (t *Table) Boolean(name string, options ...ColumnOption) *Column {
	col := &Column{
		Name: name,
		Type: &BooleanType{},
	}
	for _, option := range options {
		option(col)
	}
	t.Columns = append(t.Columns, col)
	return col
}

// Column represents a table column
type Column struct {
	// Column name in the database
	Name string
	// SQL data type of the column
	Type ColumnType
	// Whether the column allows NULL values
	Nullable bool
	// Default value expression for the column
	Default string
	// Comment or description attached to the column
	Comment string
	// Whether the column auto-increments (like SERIAL or AUTO_INCREMENT)
	AutoIncrement bool
}

// TypeSQL returns the SQL representation of the column type
func (c *Column) TypeSQL() string {
	return c.Type.SQL()
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
