package sql_wrapper

// identifiableWrapper is used to wrap existing types to provide
// Identifiable support
type identifiableWrapper struct {
	ID     int
	object Readable

	schema *schema
}

func (i *identifiableWrapper) GetID() int {
	return i.ID
}

func (i *identifiableWrapper) SetID(id int) {
	i.ID = id
}

// Object gets the encapsulated object from the wrapper
func (i identifiableWrapper) Object() Readable {
	return i.object
}

// Create a new IdentifiableWrapper object
func newIdentifiableWrapper(s *schema, object Readable, id int) identifiableWrapper {
	i := identifiableWrapper{}
	i.ID = id
	i.object = object
	i.schema = s

	return i
}
