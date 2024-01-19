package sql_wrapper

import (
	"fmt"
	"reflect"
	"strings"
)

/** TODO
- Add inference to types
- Add foreign key
*/

// AddForeignKey creates a string that adds a foreign key to another table

// SelectSQL creates a string that will select all objects in the SQL table
func (s *Schema[T]) SelectSQL() (string, error) {
	if s.table == "" {
		return "", fmt.Errorf("cannot insert record with no table name")
	}

	return fmt.Sprintf("SELECT * FROM %v;", s.table), nil
}

// DeleteSQL creates a string that will remove an object in the SQL table
func (s *Schema[T]) DeleteSQL(id int) (string, error) {
	if s.table == "" {
		return "", fmt.Errorf("cannot insert record with no table name")
	}

	return fmt.Sprintf("DELETE FROM %v WHERE id = %v;", s.table, id), nil
}

// UpdateSQL creates a string that will update the object in the SQL table
func (s *Schema[T]) UpdateSQL(id int, obj T) (string, error) {
	if s.table == "" {
		return "", fmt.Errorf("cannot insert record with no table name")
	}

	header := fmt.Sprintf("UPDATE %v SET ", s.table)
	footer := fmt.Sprintf(" WHERE id = %v;", id)
	body := ""

	// Generate the body to set columns to new values
	t := reflect.TypeOf(s.template)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		// Get the name of the field
		name, err := getName(t.Field(i))
		if err != nil {
			return "", err
		} else if name == "-" {
			// Skip fields with names '-'
			continue
		}

		// Get the value of the field
		val := v.Field(i).Interface()

		body += fmt.Sprintf("%v = %#v, ", name, val)
	}

	return header + body[0:len(body)-2] + footer, nil
}

// InsertSQL creates a string that will insert the given object into an SQL table
func (s *Schema[T]) InsertSQL(id int, obj T) (string, error) {
	// Make sure the table name is set
	if s.table == "" {
		return "", fmt.Errorf("cannot insert record with no table name")
	}

	vals := ""

	// Generate the values to insert
	t := reflect.TypeOf(s.template)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		// Skip empty fields
		if t.Field(i).Tag.Get("sql") == "-" {
			continue
		}

		// Add the value of the field
		vals += fmt.Sprintf("%#v, ", v.Field(i).Interface())
	}

	return fmt.Sprintf("INSERT INTO %v (id, %v) VALUES (%v, %v);", s.table, strings.Join(s.cols, ", "), id, vals[0:len(vals)-2]), nil
}

// CreateTableSQL creates a string that will create an SQL table
func (s *Schema[T]) CreateTableSQL() (string, error) {
	s.table = reflect.TypeOf(s.template).Name()

	// The header and footer for the create statement
	header := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %v(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, ", reflect.TypeOf(s.template).Name())
	footer := ");"
	body := ""

	// Loop through struct tags
	fields := reflect.VisibleFields(reflect.TypeOf(s.template))
	for _, field := range fields {
		// Get the name of the field
		name, err := getName(field)
		if err != nil {
			return "", err
		} else if name == "-" {
			// Skip fields with names '-'
			continue
		}
		s.cols = append(s.cols, name)

		// Get the definition of the field
		def, err := getDefinition(field)
		if err != nil {
			return "", err
		}

		body += fmt.Sprintf("%v %v, ", name, def)
	}

	return header + body[0:len(body)-2] + footer, nil
}

// getName is a helper method that gets the name of the SQL field
func getName(field reflect.StructField) (string, error) {
	n := field.Name

	// If a custom name is present, use it
	val, ok := field.Tag.Lookup("sql")
	if ok {
		n = val
	}

	// Make sure name is valid
	if n == "" {
		return n, fmt.Errorf("name (%v) is invalid", n)
	}
	return n, nil
}

// getDefinition is a helper method that gets the definition of the SQL field
func getDefinition(field reflect.StructField) (string, error) {
	// If there is a custom type, use it
	val, ok := field.Tag.Lookup("def")
	if ok {
		return val, nil
	}
	return val, fmt.Errorf("tag 'def' is not present for field '%v'", field.Name)
}
