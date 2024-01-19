package sql_wrapper

import (
	"database/sql"
	"fmt"
)

// Readable makes sure an object knows how to read itself in from SQL
// (This is a temporary and easier method to read items, I don't know the right golang shennanigans)
type Readable interface {
	Read(*sql.Rows) (map[int]Readable, error)
}

// Schema represents the build surrounding a table in SQL
type Schema[T Readable] struct {
	template T                              // The golang object to represent
	objects  map[int]identifiableWrapper[T] // Objects saved into the table
	db       *sql.DB                        // SQL Database that holds storage for the library

	table  string   // The table name
	cols   []string // Column names
	nextID int      // The next ID to set an object to
}

// Save makes sure an object is registered to the schema
func (s *Schema[T]) Save(val *T) error {
	_, err := s.validate(val)
	if err != nil {
		// If there is an error, then the object is not present and needs to be inserted
		// Object is not present, so insert it
		_, err := s.Insert(val)
		return err
	}

	// Otherwise, update the object
	return s.Update(val)
}

// Get gets the objects currently loaded
func (s *Schema[T]) Get() (map[int]*T, error) {
	// Make sure the table has a name
	if s.table == "" {
		return nil, fmt.Errorf("cannot insert record with no table name")
	}

	// Return the map
	copy := make(map[int]*T)
	for k, v := range s.objects {
		copy[k] = v.Object()
	}
	return copy, nil
}

// Insert inserts a new entry and returns the ID of the new entry
func (s *Schema[T]) Insert(val *T) (int, error) {
	// Start a transaction in the database
	tx, err := s.db.Begin()
	if err != nil {
		return -1, err
	}

	// Rollback the transaction if there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Add the object to the internal map
	id := s.nextID
	s.nextID++

	obj := newIdentifiableWrapper[T](s, val, id)

	s.objects[id] = obj

	// Add the object to SQL
	str, err := s.InsertSQL(id, *val)
	if err != nil {
		return id, err
	}

	_, err = tx.Exec(str)
	if err != nil {
		return id, err
	}

	return id, tx.Commit()
}

// Update updates an entry and returns the old object
func (s *Schema[T]) Update(val *T) error {
	obj, err := s.validate(val)
	if err != nil {
		return err
	}

	if obj.GetID() < 0 {
		return fmt.Errorf("object does not have valid id")
	}

	// Start a transaction in the database
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Rollback the transaction if there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Add the object to SQL
	str, err := s.UpdateSQL(obj.GetID(), *obj.Object())
	if err != nil {
		return err
	}

	_, err = tx.Exec(str)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Delete deletes an entry
func (s *Schema[T]) Delete(val *T) error {
	obj, err := s.validate(val)
	if err != nil {
		return err
	}

	if obj.GetID() < 0 {
		return fmt.Errorf("object does not have valid id")
	}

	// Start a transaction in the database
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Rollback the transaction if there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			// Remove the object to the internal map
			delete(s.objects, obj.GetID())
		}
	}()

	// Add the object to SQL
	str, err := s.DeleteSQL(obj.GetID())
	if err != nil {
		return err
	}

	_, err = tx.Exec(str)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Read reads an existing SQL table to populate the schema
func (s *Schema[T]) Read() error {
	// Get the entries in the database
	str, err := s.SelectSQL()
	if err != nil {
		return err
	}

	rows, err := s.db.Query(str)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Add the entries to the schema
	items, err := s.template.Read(rows)
	if err != nil {
		return err
	}

	// Loop through items and add them to the schema
	for id, val := range items {
		item, ok := val.(T)
		if !ok {
			return fmt.Errorf("items are not the correct type to save to schema")
		}

		s.objects[id] = newIdentifiableWrapper[T](s, &item, id)

		if id > s.nextID {
			s.nextID = id
		}
	}

	s.nextID++

	return nil
}

// validate is a helper method to validate that an object is a part of the schema
func (s *Schema[T]) validate(val *T) (identifiableWrapper[T], error) {
	for _, v := range s.objects {
		if val == v.Object() {
			return v, nil
		}
	}
	return identifiableWrapper[T]{}, fmt.Errorf("object is not in schema")
}

// Create a new Schema
func NewSchema[T Readable](db *sql.DB, template T) (*Schema[T], error) {
	s := &Schema[T]{db: db, template: template}
	s.objects = make(map[int]identifiableWrapper[T])
	s.nextID = 1

	// Start a transaction in the database
	tx, err := s.db.Begin()
	if err != nil {
		return s, err
	}

	// Rollback the transaction if there is an error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Create the SQL table this schema needs
	str, err := s.CreateTableSQL()
	if err != nil {
		return s, err
	}

	_, err = tx.Exec(str)
	if err != nil {
		return s, err
	}

	return s, tx.Commit()
}
