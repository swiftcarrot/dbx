package schema

// TODO: support extension version
// EnableExtension adds an extension to the schema
func (s *Schema) EnableExtension(name string) {
	for _, ext := range s.Extensions {
		if ext == name {
			return // Extension already enabled
		}
	}
	s.Extensions = append(s.Extensions, name)
}

// DisableExtension removes an extension from the schema
func (s *Schema) DisableExtension(name string) {
	for i, ext := range s.Extensions {
		if ext == name {
			s.Extensions = append(s.Extensions[:i], s.Extensions[i+1:]...)
			return
		}
	}
}
