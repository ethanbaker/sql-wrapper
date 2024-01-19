package sql_wrapper_test

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sql_wrapper "github.com/ethanbaker/sql-wrapper"
	"github.com/stretchr/testify/assert"
)

const TestObject_tableSQL = "CREATE TABLE IF NOT EXISTS TestObject(id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY, Name VARCHAR(128), Age INT(255), Weather ENUM('Summer', 'Autumn', 'Winter', 'Spring') NOT NULL);"
const TestObject_insertSQL = "INSERT INTO TestObject (id, Name, Age, Weather) VALUES (%#v, %#v, %#v, %#v);"
const TestObject_updateSQL = "UPDATE TestObject SET Name = %#v, Age = %#v, Weather = %#v WHERE id = %#v;"
const TestObject_deleteSQL = "DELETE FROM TestObject WHERE id = %#v;"

// NewMock creates a new mock SQL database for testing
func newMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func TestSave(t *testing.T) {
	assert := assert.New(t)
	db, mock := newMock()

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Expect the table creation
	mock.ExpectBegin()
	mock.ExpectExec(TestObject_tableSQL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create the new schema
	schema, err := sql_wrapper.NewSchema[TestObject](db, TestObject{})
	assert.Nil(err)

	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_insertSQL, 1, obj.Name, obj.Age, obj.Weather)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Save the test object
	err = schema.Save(&obj)
	assert.Nil(err)

	// Get the objects that are present
	objs, err := schema.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.True(reflect.DeepEqual(&obj, objs[1]))

	// Change the test object
	obj.Name = "John"

	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_updateSQL, obj.Name, obj.Age, obj.Weather, 1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = schema.Save(&obj)
	assert.Nil(err)

	// Test getting the updated object
	objs, err = schema.Get()
	assert.Nil(err)
	assert.NotNil(objs)
	assert.Equal(1, len(objs))
	assert.Equal("John", objs[1].Name)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)
	db, mock := newMock()

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Expect the table creation when creating a new schema
	mock.ExpectBegin()
	mock.ExpectExec(TestObject_tableSQL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	schema, err := sql_wrapper.NewSchema[TestObject](db, TestObject{})
	assert.Nil(err)

	// Try getting the objects when there are no objects present
	objs, err := schema.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Expect insert statement when inserting an object
	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_insertSQL, 1, obj.Name, obj.Age, obj.Weather)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := schema.Insert(&obj)
	assert.Nil(err)
	assert.Equal(1, id)

	// Get the objects that are present
	objs, err = schema.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.True(reflect.DeepEqual(&obj, objs[id]))
}

func TestInsert(t *testing.T) {
	assert := assert.New(t)
	db, mock := newMock()

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Expect the table creation when creating a new schema
	mock.ExpectBegin()
	mock.ExpectExec(TestObject_tableSQL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	schema, err := sql_wrapper.NewSchema[TestObject](db, TestObject{})
	assert.Nil(err)

	// Expect insert statement when inserting an object
	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_insertSQL, 1, obj.Name, obj.Age, obj.Weather)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := schema.Insert(&obj)
	assert.Nil(err)
	assert.Equal(1, id)
}

func TestUpdate(t *testing.T) {
	assert := assert.New(t)
	db, mock := newMock()

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Expect the table creation when creating a new schema
	mock.ExpectBegin()
	mock.ExpectExec(TestObject_tableSQL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	schema, err := sql_wrapper.NewSchema[TestObject](db, TestObject{})
	assert.Nil(err)

	// Try getting the objects when there are no objects present
	objs, err := schema.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Expect insert statement when inserting an object
	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_insertSQL, 1, obj.Name, obj.Age, obj.Weather)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := schema.Insert(&obj)
	assert.Nil(err)
	assert.Equal(1, id)

	// Get the objects that are present
	objs, err = schema.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.True(reflect.DeepEqual(&obj, objs[id]))

	// Expect an update statement when updating an object
	obj.Name = "John"

	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_updateSQL, obj.Name, obj.Age, obj.Weather, 1)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = schema.Update(&obj)
	assert.Nil(err)

	// Test getting the updated object
	objs, err = schema.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal("John", objs[id].Name)
}

func TestDelete(t *testing.T) {
	assert := assert.New(t)
	db, mock := newMock()

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Expect the table creation when creating a new schema
	mock.ExpectBegin()
	mock.ExpectExec(TestObject_tableSQL).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	schema, err := sql_wrapper.NewSchema[TestObject](db, TestObject{})
	assert.Nil(err)

	// Try getting the objects when there are no objects present
	objs, err := schema.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Expect insert statement when inserting an object
	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_insertSQL, 1, obj.Name, obj.Age, obj.Weather)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	id, err := schema.Insert(&obj)
	assert.Nil(err)
	assert.Equal(1, id)

	// Get the objects that are present
	objs, err = schema.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.True(reflect.DeepEqual(&obj, objs[id]))

	// Delete the object and expect deletion from the SQL driver
	mock.ExpectBegin()
	mock.ExpectExec(fmt.Sprintf(TestObject_deleteSQL, id)).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = schema.Delete(&obj)
	assert.Nil(err)

	// Try getting the objects when there are no objects present
	objs, err = schema.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))
}
