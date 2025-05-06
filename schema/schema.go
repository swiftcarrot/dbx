package schema

// Schema represents a database schema
type Schema struct {
	Name       string // Name of the database schema (e.g., public)
	Tables     []*Table
	Extensions []string    // PostgreSQL extensions to enable
	Sequences  []*Sequence // Database sequences
	Functions  []*Function // Database functions
	Triggers   []*Trigger  // Database triggers
	Views      []*View     // Database views
}

// NewSchema creates a new database schema definition
func NewSchema() *Schema {
	return &Schema{
		Tables:     []*Table{},
		Extensions: []string{},
		Sequences:  []*Sequence{},
		Functions:  []*Function{},
		Triggers:   []*Trigger{},
		Views:      []*View{},
	}
}

// SchemaOption represents an option for creating a schema
type SchemaOption func(*Schema)

// WithSchemaName sets the schema name for a Schema
func WithSchemaName(name string) SchemaOption {
	return func(s *Schema) {
		s.Name = name
	}
}

// CreateTable adds a new table to the schema with optional schema name
func (s *Schema) CreateTable(name string, fn func(*Table)) *Table {
	table := &Table{
		Name:    name,
		Columns: []*Column{},
		Indexes: []*Index{},
	}

	if fn != nil {
		fn(table)
	}

	s.Tables = append(s.Tables, table)
	return table
}

// CreateView adds a new view to the schema
func (s *Schema) CreateView(name string, definition string, options ...ViewOption) *View {
	view := &View{
		Name:       name,
		Schema:     s.Name,
		Definition: definition,
	}

	for _, option := range options {
		option(view)
	}

	s.Views = append(s.Views, view)
	return view
}

// CreateFunction adds a new function to the schema
func (s *Schema) CreateFunction(name string, returns string, body string, options ...FunctionOption) *Function {
	function := &Function{
		Name:       name,
		Schema:     s.Name,
		Returns:    returns,
		Body:       body,
		Language:   "plpgsql",  // Default language
		Volatility: "VOLATILE", // Default volatility
		Security:   "INVOKER",  // Default security
		Cost:       100,        // Default cost
		Arguments:  []FunctionArg{},
	}

	for _, option := range options {
		option(function)
	}

	s.Functions = append(s.Functions, function)
	return function
}

// CreateTrigger adds a new trigger to the schema
func (s *Schema) CreateTrigger(name string, tableName string, function string, options ...TriggerOption) *Trigger {
	trigger := &Trigger{
		Name:      name,
		Schema:    s.Name,
		Table:     tableName,
		Function:  function,
		Timing:    "BEFORE",           // Default timing
		Events:    []string{"INSERT"}, // Default event
		ForEach:   "ROW",              // Default scope
		Arguments: []string{},
	}

	for _, option := range options {
		option(trigger)
	}

	s.Triggers = append(s.Triggers, trigger)
	return trigger
}

// CreateSequence adds a new sequence to the schema
func (s *Schema) CreateSequence(name string, options ...SequenceOption) *Sequence {
	seq := &Sequence{
		Name:      name,
		Start:     1,
		Increment: 1,
		MinValue:  1,
		MaxValue:  9223372036854775807, // Default max for bigint
		Cache:     1,
		Cycle:     false,
	}

	for _, option := range options {
		option(seq)
	}

	s.Sequences = append(s.Sequences, seq)
	return seq
}
