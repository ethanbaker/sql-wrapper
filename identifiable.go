package sql_wrapper

// identifiableWrapper is used to wrap existing types to provide
// Identifiable support
type identifiableWrapper[T Readable] struct {
	ID     int
	object *T

	schema *Schema[T]
}

func (i *identifiableWrapper[T]) GetID() int {
	return i.ID
}

func (i *identifiableWrapper[T]) SetID(id int) {
	i.ID = id
}

// Object gets the encapsulated object from the wrapper
func (i identifiableWrapper[T]) Object() *T {
	return i.object
}

// Create a new IdentifiableWrapper object
func newIdentifiableWrapper[T Readable](schema *Schema[T], object *T, id int) identifiableWrapper[T] {
	i := identifiableWrapper[T]{}
	i.ID = id
	i.object = object
	i.schema = schema

	return i
}
