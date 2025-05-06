package schema

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Name       string
	Columns    []string
	RefTable   string
	RefColumns []string
	OnDelete   string
	OnUpdate   string
}

// ForeignKeyOption is a function type for foreign key options
type ForeignKeyOption func(*ForeignKey)

// OnDelete sets the on delete action for a foreign key
func OnDelete(action string) ForeignKeyOption {
	return func(fk *ForeignKey) {
		fk.OnDelete = action
	}
}

// OnUpdate sets the on update action for a foreign key
func OnUpdate(action string) ForeignKeyOption {
	return func(fk *ForeignKey) {
		fk.OnUpdate = action
	}
}
