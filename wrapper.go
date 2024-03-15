package sql_wrapper

import (
	"database/sql"
	"fmt"
)

// Wrapper wraps around a schema so you can call functions with defined types
type Wrapper[T Readable] struct {
	schema *schema
}

// Name returns the name of the table the schema represents
func (w *Wrapper[T]) Name() string {
	return w.schema.name()
}

// Save makes sure an object is registered to the schema and returns its ID
func (w *Wrapper[T]) Save(val T) error {
	return w.schema.Save(val)
}

// Get gets the objects currently loaded
func (w *Wrapper[T]) Get() (map[int]T, error) {
	copy := make(map[int]T)

	// Get the schema results
	result, err := w.schema.get()
	if err != nil {
		return copy, err
	}

	// Add the schema values to a list of custom types
	for k, v := range result {
		obj, ok := v.Object().(T)
		if !ok {
			return copy, fmt.Errorf("cannot cast object in schema to custom type")
		}

		copy[k] = obj

	}
	return copy, nil
}

// GetID gets an objects ID
func (w *Wrapper[T]) GetID(val T) (int, error) {
	return w.schema.getID(val)
}

// GetByID gets an object by ID
func (w *Wrapper[T]) GetByID(id int) (T, error) {
	var obj T

	val, err := w.schema.getByID(id)
	if err != nil {
		return obj, err
	}

	// Cast the object to the generic type and return
	obj, ok := val.(T)
	if !ok {
		return obj, fmt.Errorf("cannot cast object with given id to custom type")
	}

	return obj, nil
}

// Insert inserts a new entry and returns the ID of the new entry
func (w *Wrapper[T]) Insert(val T) (int, error) {
	return w.schema.insert(val)
}

// Update updates an entry and returns the old object
func (w *Wrapper[T]) Update(val T) error {
	return w.schema.update(val)
}

// Delete deletes an entry
func (w *Wrapper[T]) Delete(val T) error {
	return w.schema.delete(val)
}

// Read reads an existing SQL table to populate the schema
func (w *Wrapper[T]) Read() error {
	return w.schema.read()
}

// Create a new Schema
func NewWrapper[T Readable](db *sql.DB, template Readable) (*Wrapper[T], error) {
	w := Wrapper[T]{}

	schema, err := newSchema(db, template)
	if err != nil {
		return &w, err
	}

	w.schema = schema
	return &w, nil
}
