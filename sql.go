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

// selectSQL creates a string that will select all objects in the SQL table
/*
func (s *schema) selectSQL() (string, error) {
	if s.table == "" {
		return "", fmt.Errorf("cannot insert record with no table name")
	}

	return fmt.Sprintf("SELECT * FROM %v;", s.table), nil
}
*/

// deleteSQL creates a string that will remove an object in the SQL table
func (s *schema) deleteSQL(id int) ([]string, error) {
	statements := []string{}

	if s.table == "" {
		return statements, fmt.Errorf("cannot insert record with no table name")
	}

	// Add another statement if a one-to-many relationship is present
	t := reflect.TypeOf(s.template)
	for i := 0; i < t.NumField(); i++ {
		rel := getRelation(t.Field(i))
		if rel == OneToMany || rel == ManyToMany {
			// Get the combined table name
			tableRef := strings.Split(t.Field(i).Type.String(), ".")[1]
			combinedTable := s.table + tableRef

			statements = append(statements, fmt.Sprintf("DELETE FROM %v WHERE %vID = %v", combinedTable, s.table, id))
		}
	}

	statements = append(statements, fmt.Sprintf("DELETE FROM %v WHERE id = %v;", s.table, id))
	return statements, nil
}

// updateSQL creates a string that will update the object in the SQL table
func (s *schema) updateSQL(id int, obj Readable) ([]string, error) {
	statements := []string{}

	if s.table == "" {
		return statements, fmt.Errorf("cannot insert record with no table name")
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
			// Stop on error
			return statements, err
		} else if name == "-" {
			// Skip fields with names '-'
			continue
		}

		// Determine if the field is a foreign relation
		rel := getRelation(t.Field(i))
		if rel == UndefinedRelationType {
			// Attribute is not a foreign relation so add normally
			val := v.Elem().Field(i).Interface()

			body += fmt.Sprintf("%v = %#v, ", name, val)
		} else if rel == OneToOne || rel == ManyToOne {
			// In the case of OneToOne or ManyToOne relationships, update to the ID to the field
			tableRef := t.Field(i).Type.Elem().Name()

			// Get the schema the object belongs to
			schema, err := manager.getSchema(tableRef)
			if err != nil {
				return statements, err
			}

			// Dereference the object that implements the Readable Interface
			obj, ok := v.Elem().Field(i).Interface().(Readable)
			if !ok {
				return statements, fmt.Errorf("cannot cast schema object as Readable")
			}

			if obj != nil && !reflect.ValueOf(obj).IsNil() {
				// Get the ID of the object and add it if the object is not nil
				id, err := schema.getID(obj)
				if err != nil {
					return statements, err
				}

				body += fmt.Sprintf("%v = %#v, ", name, id)
			} else {
				// If the object is nil, insert null
				body += fmt.Sprintf("%v = NULL, ", name)
			}
		} else if rel == OneToMany || rel == ManyToMany {
			// In the case of a OneToMany or ManyToMany relationship, update entries to another table
			tableRef := strings.Split(t.Field(i).Type.String(), ".")[1]
			combinedTable := s.table + tableRef

			// Delete entries that previously exist with this ID
			statements = append(statements, fmt.Sprintf("DELETE FROM %v WHERE %vID = %v;", combinedTable, s.table, id))

			// Get the list of objects
			slice := v.Elem().Field(i)
			if slice.Kind() != reflect.Slice {
				return statements, fmt.Errorf("relationship does not have slice type")
			}

			// Get the schema
			schema, err := manager.getSchema(tableRef)
			if err != nil {
				return statements, err
			}

			for i := 0; i < slice.Len(); i++ {
				val := slice.Index(i)

				// Cast the val to a Readable object
				readable, ok := val.Interface().(Readable)
				if !ok {
					return statements, fmt.Errorf("cannot cast element in relationship to Readable")
				}

				// Get the ID of the object
				objID, err := schema.getID(readable)
				if err != nil {
					return statements, err
				}

				statements = append(statements, fmt.Sprintf("INSERT INTO %v(%v, %vID) VALUES (%v, %v);", combinedTable, name, s.table, objID, id))
			}
		}
	}

	statements = append([]string{header + body[0:len(body)-2] + footer}, statements...)

	return statements, nil
}

// insertSQL creates a string that will insert the given object into an SQL table
func (s *schema) insertSQL(id int, obj Readable) ([]string, error) {
	statements := []string{}

	// Make sure the table name is set
	if s.table == "" {
		return statements, fmt.Errorf("cannot insert record with no table name")
	}

	header := fmt.Sprintf("INSERT INTO %v (id, ", s.table)
	body := fmt.Sprintf(") VALUES (%v, ", id)
	footer := ");"

	var columns []string

	// Generate the values to insert
	t := reflect.TypeOf(s.template)
	v := reflect.ValueOf(obj)

	// Create an insert string for normal attributes
	for i := 0; i < t.NumField(); i++ {
		// Get the name of the field
		name, err := getName(t.Field(i))
		if err != nil {
			// Stop on error
			return statements, err
		} else if name == "-" {
			// Skip fields with names '-'
			continue
		}

		// Determine if the field is a foreign relation
		rel := getRelation(t.Field(i))
		if rel == UndefinedRelationType {
			// Attribute is not a foreign relation so add normally
			columns = append(columns, name)
			body += fmt.Sprintf("%#v, ", v.Elem().Field(i).Interface())
		} else if rel == OneToOne || rel == ManyToOne {
			// Attribute is a one-to-one or many-to-one foreign relation
			tableRef := t.Field(i).Type.Elem().Name()

			// In the case of OneToOne or ManyToOne relationships, add the ID to the field
			columns = append(columns, name)

			// Get the schema the object belongs to
			schema, err := manager.getSchema(tableRef)
			if err != nil {
				return statements, err
			}

			// Dereference the object that implements the Readable Interface
			val, ok := v.Elem().Field(i).Interface().(Readable)
			if !ok {
				return statements, fmt.Errorf("cannot cast schema object as Readable")
			}

			if val != nil && !reflect.ValueOf(val).IsNil() {
				// Get the ID of the object if it is not nil
				id, err := schema.getID(val)
				if err != nil {
					return statements, err
				}

				body += fmt.Sprintf("%#v, ", id)
			} else {
				// If the object is nil, insert null
				body += "NULL, "
			}
		} else if rel == OneToMany || rel == ManyToMany {
			// In the case of a OneToMany or ManyToMany relationships, add entries to another table
			tableRef := strings.Split(t.Field(i).Type.String(), ".")[1]
			combinedTable := s.table + tableRef

			// Get the list of objects
			slice := v.Elem().Field(i)
			if slice.Kind() != reflect.Slice {
				return statements, fmt.Errorf("relationship does not have slice type")
			}

			// Get the schema
			schema, err := manager.getSchema(tableRef)
			if err != nil {
				return statements, err
			}

			for i := 0; i < slice.Len(); i++ {
				val := slice.Index(i)

				// Cast the val to a Readable object
				readable, ok := val.Interface().(Readable)
				if !ok {
					return statements, fmt.Errorf("cannot cast element in relationship to Readable")
				}

				// Get the ID of the object
				objID, err := schema.getID(readable)
				if err != nil {
					return statements, err
				}

				statements = append(statements, fmt.Sprintf("INSERT INTO %v(%v, %vID) VALUES (%v, %v);", combinedTable, name, s.table, objID, id))
			}
		}
	}

	statements = append(statements, header+strings.Join(columns, ", ")+body[0:len(body)-2]+footer)

	temp := statements[0]
	statements[0] = statements[len(statements)-1]
	statements[len(statements)-1] = temp

	return statements, nil
}

// createTableSQL creates a string that will create an SQL table
func (s *schema) createTableSQL() ([]string, error) {
	statements := []string{}

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
			// Stop on error
			return statements, err
		} else if name == "-" {
			// Skip fields with names '-'
			continue
		}

		// Determine if the field is a foreign relation
		rel := getRelation(field)
		if rel == UndefinedRelationType {
			// Field is not a foreign relation so add normally
			s.cols = append(s.cols, name)

			// Get the definition of the field
			def, err := getDefinition(field)
			if err != nil {
				return statements, err
			}

			// Add the name and definition to the SQL
			body += fmt.Sprintf("%v %v, ", name, def)
		} else if rel == OneToOne {
			// The field has a one-to-one foreign relation
			tableRef := field.Type.Elem().Name()

			body += fmt.Sprintf("%[1]v INT UNSIGNED UNIQUE, FOREIGN KEY (%[1]v) REFERENCES %[2]v(id) ON DELETE CASCADE ON UPDATE CASCADE, ", name, tableRef)
		} else if rel == ManyToOne {
			// The field has a many-to-one foreign relation
			tableRef := field.Type.Elem().Name()

			body += fmt.Sprintf("%[1]v INT UNSIGNED, FOREIGN KEY (%[1]v) REFERENCES %[2]v(id) ON DELETE CASCADE ON UPDATE CASCADE, ", name, tableRef)
		} else if rel == OneToMany {
			// The field is a one-to-many foreign relation
			tableRef := strings.Split(field.Type.String(), ".")[1]

			statements = append(statements, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %[1]v%[2]v(%[1]vID INT UNSIGNED, %[3]v INT UNSIGNED UNIQUE, FOREIGN KEY (%[1]vID) REFERENCES %[1]v(id), FOREIGN KEY (%[3]v) REFERENCES %[2]v(id));", s.table, tableRef, name))
		} else if rel == ManyToMany {
			// The field is a many-to-many foreign relation
			tableRef := strings.Split(field.Type.String(), ".")[1]

			statements = append(statements, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %[1]v%[2]v(%[1]vID INT UNSIGNED, %[3]v INT UNSIGNED, FOREIGN KEY (%[1]vID) REFERENCES %[1]v(id), FOREIGN KEY (%[3]v) REFERENCES %[2]v(id));", s.table, tableRef, name))
		}
	}

	statements = append(statements, header+body[0:len(body)-2]+footer)

	temp := statements[0]
	statements[0] = statements[len(statements)-1]
	statements[len(statements)-1] = temp

	return statements, nil
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
	val, ok := field.Tag.Lookup("def")
	if ok {
		return val, nil
	}
	return val, fmt.Errorf("tag 'def' is not present for field '%v'", field.Name)
}

// getRelation is a helper method that gets the relation type of the SQL field
func getRelation(field reflect.StructField) RelationType {
	val, ok := field.Tag.Lookup("rel")
	if !ok {
		return UndefinedRelationType
	}

	switch val {
	case "one-to-one":
		return OneToOne

	case "one-to-many":
		return OneToMany

	case "many-to-one":
		return ManyToOne

	case "many-to-many":
		return ManyToMany

	default:
		return UndefinedRelationType
	}
}
