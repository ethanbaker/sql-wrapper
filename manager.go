package sql_wrapper

import "fmt"

// schemaManager manages multiple schemas together and handles foreign references
type schemaManager struct {
	schemas map[string]*schema
}

// addSchema adds a schema to the schemaManager
func (m *schemaManager) addSchema(s interface{}) error {
	// Initialize map of schemas if not present
	if m.schemas == nil {
		m.schemas = make(map[string]*schema)
	}

	// Cast the interface to a schema
	readableSchema, ok := s.(*schema)
	if !ok {
		return fmt.Errorf("cannot cast interface into readable schema")
	}

	m.schemas[readableSchema.name()] = readableSchema

	return nil
}

// getSchema returns a schema with a given name
func (m *schemaManager) getSchema(name string) (*schema, error) {
	schema, ok := m.schemas[name]
	if !ok {
		return nil, fmt.Errorf("schema is not in schemaManager")
	}

	return schema, nil
}

// manager holds all schemas locally so they can reference one another
var manager schemaManager

// GetObjectBySchema is used by read methods to get objects in other schemas
func GetObjectBySchema(name string, id int) (Readable, error) {
	// Get the schema from the manager
	schema, err := manager.getSchema(name)
	if err != nil {
		return nil, err
	}

	// Get the object
	return schema.getByID(id)
}
