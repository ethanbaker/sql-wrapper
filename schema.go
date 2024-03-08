package sql_wrapper

import (
	"database/sql"
	"fmt"
)

// Readable makes sure an object knows how to read itself in from SQL
// (This is a temporary and easier method to read items, I don't know the right golang shennanigans)
type Readable interface {
	Read(*sql.DB) (map[int]Readable, error)
}

// Schema represents the build surrounding a table in SQL
type schema struct {
	template Readable                    // The golang object to represent
	objects  map[int]identifiableWrapper // Objects saved into the table
	db       *sql.DB                     // SQL Database that holds storage for the library

	table  string   // The table name
	cols   []string // Column names
	nextID int      // The next ID to set an object to
}

// name returns the name of the table the schema represents
func (s *schema) name() string {
	return s.table
}

// save makes sure an object is registered to the schema and returns its ID
func (s *schema) Save(val Readable) error {
	_, err := s.validate(val)
	if err != nil {
		// If there is an error, then the object is not present and needs to be inserted
		// Object is not present, so insert it
		_, err := s.insert(val)
		return err
	}

	// Otherwise, update the object
	return s.update(val)
}

// get gets the objects currently loaded
func (s *schema) get() (map[int]identifiableWrapper, error) {
	// Make sure the table has a name
	if s.table == "" {
		return nil, fmt.Errorf("cannot insert record with no table name")
	}

	// Return the map
	return s.objects, nil
}

// getID gets an objects ID
func (s *schema) getID(val Readable) (int, error) {
	obj, err := s.validate(val)
	if err != nil {
		return -1, err
	}

	id := obj.GetID()
	if id < 0 {
		return -1, fmt.Errorf("object does not have valid id")
	}

	return id, nil
}

// getByID gets an object from its ID
func (s *schema) getByID(id int) (Readable, error) {
	obj, ok := s.objects[id]
	if !ok {
		return nil, fmt.Errorf("no object with id in schema")
	}

	return obj.Object(), nil
}

// insert inserts a new entry and returns the ID of the new entry
func (s *schema) insert(val Readable) (int, error) {
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

	obj := newIdentifiableWrapper(s, val, id)

	s.objects[id] = obj

	// Add the object to SQL
	strs, err := s.insertSQL(id, val)
	if err != nil {
		return id, err
	}

	for _, str := range strs {
		_, err = tx.Exec(str)
		if err != nil {
			return id, err
		}
	}

	return id, tx.Commit()
}

// update updates an entry and returns the old object
func (s *schema) update(val Readable) error {
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
	strs, err := s.updateSQL(obj.GetID(), obj.Object())
	if err != nil {
		return err
	}

	for _, str := range strs {
		_, err = tx.Exec(str)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// delete deletes an entry
func (s *schema) delete(val Readable) error {
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
	strs, err := s.deleteSQL(obj.GetID())
	if err != nil {
		return err
	}

	for _, str := range strs {
		_, err = tx.Exec(str)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// read reads an existing SQL table to populate the schema
func (s *schema) read() error {
	// Add the entries to the schema
	items, err := s.template.Read(s.db)
	if err != nil {
		return err
	}

	// Loop through items and add them to the schema
	s.nextID = 0
	for id, val := range items {
		s.objects[id] = newIdentifiableWrapper(s, val, id)

		if id > s.nextID {
			s.nextID = id
		}
	}

	s.nextID++

	return nil
}

// validate is a helper method to validate that an object is a part of the schema
func (s *schema) validate(val Readable) (identifiableWrapper, error) {
	for _, v := range s.objects {
		if val == v.Object() {
			return v, nil
		}
	}
	return identifiableWrapper{}, fmt.Errorf("object is not in schema")
}

// newSchema creates a new Schema
func newSchema(db *sql.DB, template Readable) (*schema, error) {
	s := &schema{db: db, template: template}
	s.objects = make(map[int]identifiableWrapper)
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
	strs, err := s.createTableSQL()
	if err != nil {
		return s, err
	}

	for _, str := range strs {
		_, err = tx.Exec(str)
		if err != nil {
			return s, err
		}
	}

	// Add the schema to the manager
	manager.addSchema(s)

	return s, tx.Commit()
}
