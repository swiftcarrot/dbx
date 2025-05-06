package schema

// ChangeType defines the type of schema change
type ChangeType string

const (
	CreateSchema     ChangeType = "create_schema"
	DropSchema       ChangeType = "drop_schema"
	EnableExtension  ChangeType = "enable_extension"
	DisableExtension ChangeType = "disable_extension"
	CreateTable      ChangeType = "create_table"
	DropTable        ChangeType = "drop_table"
	AddColumn        ChangeType = "add_column"
	DropColumn       ChangeType = "drop_column"
	AlterColumn      ChangeType = "alter_column"
	AddPrimaryKey    ChangeType = "add_primary_key"
	DropPrimaryKey   ChangeType = "drop_primary_key"
	AddIndex         ChangeType = "add_index"
	DropIndex        ChangeType = "drop_index"
	AddForeignKey    ChangeType = "add_foreign_key"
	DropForeignKey   ChangeType = "drop_foreign_key"
	CreateSequence   ChangeType = "create_sequence"
	DropSequence     ChangeType = "drop_sequence"
	AlterSequence    ChangeType = "alter_sequence"
	CreateFunction   ChangeType = "create_function"
	AlterFunction    ChangeType = "alter_function"
	DropFunction     ChangeType = "drop_function"
	CreateView       ChangeType = "create_view"
	AlterView        ChangeType = "alter_view"
	DropView         ChangeType = "drop_view"
	CreateTrigger    ChangeType = "create_trigger"
	AlterTrigger     ChangeType = "alter_trigger"
	DropTrigger      ChangeType = "drop_trigger"
)

// Change is an interface representing a database schema change
type Change interface {
	Type() ChangeType
	IsUnsafe() bool
}

// BaseChange provides common functionality for all change types
type BaseChange struct {
	unsafe bool
}

// IsUnsafe returns whether the change is unsafe and should be reviewed before execution
func (c BaseChange) IsUnsafe() bool {
	return c.unsafe
}

// SetUnsafe marks a change as unsafe
func (c *BaseChange) SetUnsafe(unsafe bool) {
	c.unsafe = unsafe
}

// Table-related changes

// CreateTableChange represents a table creation change
type CreateTableChange struct {
	BaseChange
	TableDef *Table
}

func (c CreateTableChange) Type() ChangeType {
	return CreateTable
}

// DropTableChange represents a table drop change
type DropTableChange struct {
	BaseChange
	TableName  string
	SchemaName string
}

func (c DropTableChange) Type() ChangeType {
	return DropTable
}

// Column-related changes

// AddColumnChange represents adding a column to a table
type AddColumnChange struct {
	BaseChange
	TableName string
	Column    *Column
}

func (c AddColumnChange) Type() ChangeType {
	return AddColumn
}

// DropColumnChange represents dropping a column from a table
type DropColumnChange struct {
	BaseChange
	TableName  string
	ColumnName string
}

func (c DropColumnChange) Type() ChangeType {
	return DropColumn
}

// AlterColumnChange represents altering a column in a table
type AlterColumnChange struct {
	BaseChange
	TableName string
	Column    *Column
}

func (c AlterColumnChange) Type() ChangeType {
	return AlterColumn
}

// PrimaryKey-related changes

// AddPrimaryKeyChange represents adding a primary key to a table
type AddPrimaryKeyChange struct {
	BaseChange
	TableName  string
	PrimaryKey *PrimaryKey
}

func (c AddPrimaryKeyChange) Type() ChangeType {
	return AddPrimaryKey
}

// DropPrimaryKeyChange represents dropping a primary key from a table
type DropPrimaryKeyChange struct {
	BaseChange
	TableName string
	PKName    string
}

func (c DropPrimaryKeyChange) Type() ChangeType {
	return DropPrimaryKey
}

// Index-related changes

// AddIndexChange represents adding an index to a table
type AddIndexChange struct {
	BaseChange
	TableName string
	Index     *Index
}

func (c AddIndexChange) Type() ChangeType {
	return AddIndex
}

// DropIndexChange represents dropping an index from a table
type DropIndexChange struct {
	BaseChange
	TableName string
	IndexName string
}

func (c DropIndexChange) Type() ChangeType {
	return DropIndex
}

// ForeignKey-related changes

// AddForeignKeyChange represents adding a foreign key to a table
type AddForeignKeyChange struct {
	BaseChange
	TableName  string
	ForeignKey *ForeignKey
}

func (c AddForeignKeyChange) Type() ChangeType {
	return AddForeignKey
}

// DropForeignKeyChange represents dropping a foreign key from a table
type DropForeignKeyChange struct {
	BaseChange
	TableName string
	FKName    string
}

func (c DropForeignKeyChange) Type() ChangeType {
	return DropForeignKey
}

// Schema-related changes

// CreateSchemaChange represents creating a new schema
type CreateSchemaChange struct {
	BaseChange
	SchemaName string
}

func (c CreateSchemaChange) Type() ChangeType {
	return CreateSchema
}

// DropSchemaChange represents dropping a schema
type DropSchemaChange struct {
	BaseChange
	SchemaName string
}

func (c DropSchemaChange) Type() ChangeType {
	return DropSchema
}

// Extension-related changes

// EnableExtensionChange represents enabling a PostgreSQL extension
type EnableExtensionChange struct {
	BaseChange
	Extension string
}

func (c EnableExtensionChange) Type() ChangeType {
	return EnableExtension
}

// DisableExtensionChange represents disabling a PostgreSQL extension
type DisableExtensionChange struct {
	BaseChange
	Extension string
}

func (c DisableExtensionChange) Type() ChangeType {
	return DisableExtension
}

// Sequence-related changes

// CreateSequenceChange represents creating a new sequence
type CreateSequenceChange struct {
	BaseChange
	Sequence *Sequence
}

func (c CreateSequenceChange) Type() ChangeType {
	return CreateSequence
}

// AlterSequenceChange represents altering an existing sequence
type AlterSequenceChange struct {
	BaseChange
	Sequence *Sequence
}

func (c AlterSequenceChange) Type() ChangeType {
	return AlterSequence
}

// DropSequenceChange represents dropping a sequence
type DropSequenceChange struct {
	BaseChange
	SchemaName   string
	SequenceName string
}

func (c DropSequenceChange) Type() ChangeType {
	return DropSequence
}

// Function-related changes

// CreateFunctionChange represents creating a new function
type CreateFunctionChange struct {
	BaseChange
	Function *Function
}

func (c CreateFunctionChange) Type() ChangeType {
	return CreateFunction
}

// AlterFunctionChange represents altering an existing function
type AlterFunctionChange struct {
	BaseChange
	Function *Function
}

func (c AlterFunctionChange) Type() ChangeType {
	return AlterFunction
}

// DropFunctionChange represents dropping a function
type DropFunctionChange struct {
	BaseChange
	SchemaName   string
	FunctionName string
	FunctionArgs []FunctionArg // Needed to identify overloaded functions
}

func (c DropFunctionChange) Type() ChangeType {
	return DropFunction
}

// View-related changes

// CreateViewChange represents creating a new view
type CreateViewChange struct {
	BaseChange
	View *View
}

func (c CreateViewChange) Type() ChangeType {
	return CreateView
}

// AlterViewChange represents altering an existing view
type AlterViewChange struct {
	BaseChange
	View *View
}

func (c AlterViewChange) Type() ChangeType {
	return AlterView
}

// DropViewChange represents dropping a view
type DropViewChange struct {
	BaseChange
	SchemaName string
	ViewName   string
}

func (c DropViewChange) Type() ChangeType {
	return DropView
}

// Trigger-related changes

// CreateTriggerChange represents creating a new trigger
type CreateTriggerChange struct {
	BaseChange
	Trigger *Trigger
}

func (c CreateTriggerChange) Type() ChangeType {
	return CreateTrigger
}

// AlterTriggerChange represents altering an existing trigger
type AlterTriggerChange struct {
	BaseChange
	Trigger *Trigger
}

func (c AlterTriggerChange) Type() ChangeType {
	return AlterTrigger
}

// DropTriggerChange represents dropping a trigger
type DropTriggerChange struct {
	BaseChange
	SchemaName   string
	TriggerName  string
	TriggerTable string
}

func (c DropTriggerChange) Type() ChangeType {
	return DropTrigger
}
