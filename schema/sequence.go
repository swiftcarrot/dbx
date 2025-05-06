package schema

// Sequence represents a database sequence
type Sequence struct {
	Schema    string // Schema containing the sequence
	Name      string
	Start     int64 // Start value
	Increment int64 // Increment value
	MinValue  int64 // Minimum value
	MaxValue  int64 // Maximum value
	Cache     int64 // Cache size
	Cycle     bool  // Whether the sequence cycles when it reaches the limit
}

// SequenceOption represents an option for creating a sequence
type SequenceOption func(*Sequence)

// Start sets the starting value for a sequence
func Start(value int64) SequenceOption {
	return func(s *Sequence) {
		s.Start = value
	}
}

// Increment sets the increment value for a sequence
func Increment(value int64) SequenceOption {
	return func(s *Sequence) {
		s.Increment = value
	}
}

// MinValue sets the minimum value for a sequence
func MinValue(value int64) SequenceOption {
	return func(s *Sequence) {
		s.MinValue = value
	}
}

// MaxValue sets the maximum value for a sequence
func MaxValue(value int64) SequenceOption {
	return func(s *Sequence) {
		s.MaxValue = value
	}
}

// Cache sets the cache size for a sequence
func Cache(value int64) SequenceOption {
	return func(s *Sequence) {
		s.Cache = value
	}
}

// Cycle makes the sequence cycle when it reaches its limit
func Cycle(s *Sequence) {
	s.Cycle = true
}

// NoCycle makes the sequence not cycle when it reaches its limit
func NoCycle(s *Sequence) {
	s.Cycle = false
}

// InSchema sets the schema name for a sequence
func InSchema(name string) SequenceOption {
	return func(s *Sequence) {
		s.Schema = name
	}
}
