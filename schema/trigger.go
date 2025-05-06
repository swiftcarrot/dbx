package schema

// Trigger represents a database trigger
type Trigger struct {
	Schema    string // Schema containing the trigger
	Name      string
	Table     string   // Table the trigger is attached to
	Events    []string // INSERT, UPDATE, DELETE
	Timing    string   // BEFORE, AFTER, or INSTEAD OF
	ForEach   string   // ROW or STATEMENT
	When      string   // Optional condition
	Function  string   // Function to call
	Arguments []string // Arguments to pass to the function
}

// TriggerOption represents an option for creating a trigger
type TriggerOption func(*Trigger)

// Before sets the trigger timing to BEFORE
func Before(t *Trigger) {
	t.Timing = "BEFORE"
}

// After sets the trigger timing to AFTER
func After(t *Trigger) {
	t.Timing = "AFTER"
}

// InsteadOf sets the trigger timing to INSTEAD OF (for views)
func InsteadOf(t *Trigger) {
	t.Timing = "INSTEAD OF"
}

// ForEachRow sets the trigger to fire for each row
func ForEachRow(t *Trigger) {
	t.ForEach = "ROW"
}

// ForEachStatement sets the trigger to fire once per statement
func ForEachStatement(t *Trigger) {
	t.ForEach = "STATEMENT"
}

// OnEvents sets the events that fire the trigger
func OnEvents(events ...string) TriggerOption {
	return func(t *Trigger) {
		t.Events = events
	}
}

// WithCondition sets the WHEN condition for a trigger
func WithCondition(condition string) TriggerOption {
	return func(t *Trigger) {
		t.When = condition
	}
}

// WithArguments sets the arguments to pass to the trigger function
func WithArguments(args ...string) TriggerOption {
	return func(t *Trigger) {
		t.Arguments = args
	}
}

// TriggerInSchema sets the schema name for a trigger
func TriggerInSchema(schema string) TriggerOption {
	return func(t *Trigger) {
		t.Schema = schema
	}
}
