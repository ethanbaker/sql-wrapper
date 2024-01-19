package sql_wrapper_test

import (
	"database/sql"
	"testing"

	"github.com/ethanbaker/sql_wrapper"
	"github.com/stretchr/testify/assert"
)

// Season is used to encode an enum into SQL
type Season string

const (
	Undefined Season = "Undefined"
	Summer    Season = "Summer"
	Autumn    Season = "Autumn"
	Winter    Season = "Winter"
	Spring    Season = "Spring"
)

// TestObject is used to encode tables into SQL
type TestObject struct {
	// Generic struct attributes
	Name    string `sql:"Name" def:"VARCHAR(128)"`
	Age     int    `sql:"Age" def:"INT(255)"`
	Weather Season `def:"ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL"`
	Hidden  string `sql:"-"`
}

// Read reads in TestObjects from an SQL query
func (t TestObject) Read(rows *sql.Rows) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Read for each row
	id := 0
	name := ""
	age := 0
	weather := Undefined
	for rows.Next() {
		if err := rows.Scan(&id, &name, &age, &weather); err != nil {
			return items, err
		}

		obj := TestObject{Name: name, Age: age, Weather: weather}
		items[id] = obj
	}

	return items, nil
}

var schema sql_wrapper.Schema[TestObject]

// Test CreateTableSQL
func TestCreateTableSQL(t *testing.T) {
	assert := assert.New(t)

	schema = sql_wrapper.Schema[TestObject]{}

	str, err := schema.CreateTableSQL()
	assert.Nil(err)
	assert.Equal("CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);", str)
}

// Test InsertSQL
func TestInsertSQL(t *testing.T) {
	assert := assert.New(t)

	schema = sql_wrapper.Schema[TestObject]{}
	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Table needs to be created before insert
	_, err := schema.InsertSQL(0, obj)
	assert.NotNil(err)
	assert.Equal("cannot insert record with no table name", err.Error())

	// Create table
	str, err := schema.CreateTableSQL()
	assert.Nil(err)
	assert.Equal("CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);", str)

	// Now can insert properly
	str, err = schema.InsertSQL(0, obj)
	assert.Nil(err)
	assert.Equal(`INSERT INTO TestObject (id, Name, Age, Weather) VALUES (0, "Jack", 20, "Summer");`, str)
}

// Test UpdateSQL
func TestUpdateSQL(t *testing.T) {
	assert := assert.New(t)

	schema = sql_wrapper.Schema[TestObject]{}
	obj1 := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}
	obj2 := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Table needs to be created before update
	_, err := schema.UpdateSQL(0, obj1)
	assert.NotNil(err)
	assert.Equal("cannot insert record with no table name", err.Error())

	// Create table
	str, err := schema.CreateTableSQL()
	assert.Nil(err)
	assert.Equal("CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);", str)

	// Insert object
	str, err = schema.InsertSQL(0, obj1)
	assert.Nil(err)
	assert.Equal(`INSERT INTO TestObject (id, Name, Age, Weather) VALUES (0, "Jack", 20, "Summer");`, str)

	// Update object
	str, err = schema.UpdateSQL(0, obj2)
	assert.Nil(err)
	assert.Equal(`UPDATE TestObject SET Name = "Jack", Age = 20, Weather = "Summer" WHERE id = 0;`, str)
}

// Test DeleteSQL
func TestDelteSQL(t *testing.T) {
	assert := assert.New(t)

	schema = sql_wrapper.Schema[TestObject]{}

	// Table needs to be created before insert
	_, err := schema.DeleteSQL(0)
	assert.NotNil(err)
	assert.Equal("cannot insert record with no table name", err.Error())

	// Create table
	str, err := schema.CreateTableSQL()
	assert.Nil(err)
	assert.Equal("CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);", str)

	// Try deleting an object
	str, err = schema.DeleteSQL(0)
	assert.Nil(err)
	assert.Equal(`DELETE FROM TestObject WHERE id = 0;`, str)
}

// Test SelectSQL
func TestSelectSQL(t *testing.T) {
	assert := assert.New(t)

	schema = sql_wrapper.Schema[TestObject]{}

	// Table needs to be created before insert
	_, err := schema.SelectSQL()
	assert.NotNil(err)
	assert.Equal("cannot insert record with no table name", err.Error())

	// Create table
	str, err := schema.CreateTableSQL()
	assert.Nil(err)
	assert.Equal("CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);", str)

	// Try deleting an object
	str, err = schema.SelectSQL()
	assert.Nil(err)
	assert.Equal(`SELECT * FROM TestObject;`, str)
}
