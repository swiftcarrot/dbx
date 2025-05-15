package schema

// RowPolicy represents a PostgreSQL row level security policy
type RowPolicy struct {
	// Database schema containing the table
	Schema string
	// Name of the table the policy applies to
	TableName string
	// Name of the row level security policy
	PolicyName string
	// Can be: ALL, SELECT, INSERT, UPDATE, or DELETE
	CommandType string
	// List of roles this policy applies to
	Roles []string
	// USING expression for filtering rows visible to operations
	UsingExpr string
	// WITH CHECK expression for filtering rows that can be added
	CheckExpr string
	// true for PERMISSIVE (default), false for RESTRICTIVE
	Permissive bool
}

// RowPolicyOption represents an option for creating a row policy
type RowPolicyOption func(*RowPolicy)

// RowPolicyForCommands sets the command type for a row policy
func RowPolicyForCommands(commandType string) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.CommandType = commandType
	}
}

// RowPolicyForRoles sets the roles for a row policy
func RowPolicyForRoles(roleNames ...string) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.Roles = roleNames
	}
}

// RowPolicyUsingExpr sets the USING expression for a row policy
func RowPolicyUsingExpr(expr string) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.UsingExpr = expr
	}
}

// RowPolicyCheckExpr sets the CHECK expression for a row policy
func RowPolicyCheckExpr(expr string) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.CheckExpr = expr
	}
}

// RowPolicyPermissive sets whether the policy is permissive (default) or restrictive
func RowPolicyPermissive(permissive bool) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.Permissive = permissive
	}
}

// RowPolicyInSchema sets the schema for a row policy
func RowPolicyInSchema(schema string) RowPolicyOption {
	return func(rp *RowPolicy) {
		rp.Schema = schema
	}
}
