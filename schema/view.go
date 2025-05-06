package schema

// View represents a database view
type View struct {
	Schema     string
	Name       string
	Definition string
	Options    []string
	Columns    []string
}

// ViewOption represents an option for creating a view
type ViewOption func(*View)

// ViewColumns sets explicit column names for a view
func ViewColumns(columns ...string) ViewOption {
	return func(v *View) {
		v.Columns = columns
	}
}

// ViewOptions sets additional options for a view (e.g., security_barrier)
func ViewOptions(options ...string) ViewOption {
	return func(v *View) {
		v.Options = options
	}
}

// ViewInSchema sets the schema name for a view
func ViewInSchema(schema string) ViewOption {
	return func(v *View) {
		v.Schema = schema
	}
}
