package schema

// Function represents a database function
type Function struct {
	Schema     string
	Name       string
	Arguments  []FunctionArg
	Returns    string
	Language   string
	Body       string
	Volatility string
	Strict     bool
	Security   string
	Cost       int
}

// FunctionArg represents a function argument
type FunctionArg struct {
	Name    string
	Type    string
	Mode    string // IN, OUT, INOUT, or VARIADIC
	Default string
}

// FunctionOption represents an option for creating a function
type FunctionOption func(*Function)

// Language sets the language for a function
func Language(lang string) FunctionOption {
	return func(f *Function) {
		f.Language = lang
	}
}

// Immutable marks a function as immutable
func Immutable(f *Function) {
	f.Volatility = "IMMUTABLE"
}

// Stable marks a function as stable
func Stable(f *Function) {
	f.Volatility = "STABLE"
}

// Volatile marks a function as volatile
func Volatile(f *Function) {
	f.Volatility = "VOLATILE"
}

// Strict marks a function as strict (returns null if any argument is null)
func Strict(f *Function) {
	f.Strict = true
}

// NotStrict marks a function as not strict
func NotStrict(f *Function) {
	f.Strict = false
}

// SecurityDefiner sets the security context to the function definer
func SecurityDefiner(f *Function) {
	f.Security = "DEFINER"
}

// SecurityInvoker sets the security context to the function invoker
func SecurityInvoker(f *Function) {
	f.Security = "INVOKER"
}

// FunctionCost sets the estimated execution cost for a function
func FunctionCost(cost int) FunctionOption {
	return func(f *Function) {
		f.Cost = cost
	}
}

// FunctionInSchema sets the schema name for a function
func FunctionInSchema(schema string) FunctionOption {
	return func(f *Function) {
		f.Schema = schema
	}
}

// FunctionArgs adds arguments to a function
func FunctionArgs(args ...FunctionArg) FunctionOption {
	return func(f *Function) {
		f.Arguments = args
	}
}

// NewFunctionArg creates a new function argument
func NewFunctionArg(name string, typeName string) FunctionArg {
	return FunctionArg{
		Name: name,
		Type: typeName,
		Mode: "IN", // Default mode
	}
}

// WithMode sets the mode for a function argument
func (arg FunctionArg) WithMode(mode string) FunctionArg {
	arg.Mode = mode
	return arg
}

// WithDefault sets a default value for a function argument
func (arg FunctionArg) WithDefault(defaultValue string) FunctionArg {
	arg.Default = defaultValue
	return arg
}
