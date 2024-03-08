package sql_wrapper_test

import (
	"database/sql"
	"log"
	"testing"

	sql_wrapper "github.com/ethanbaker/sql-wrapper"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

// ---------- Types ----------

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
func (t TestObject) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM TestObject")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	for rows.Next() {
		if err := rows.Scan(&id, &name, &age, &weather); err != nil {
			return items, err
		}

		obj := TestObject{Name: name, Age: age, Weather: weather}
		items[id] = &obj
	}

	return items, nil
}

// ---------- Globals ----------

var database *sql.DB
var wrapper *sql_wrapper.Wrapper[*TestObject]

var cfg = mysql.Config{
	User:   "sql_wrapper_test",
	Passwd: "abc123",
	Net:    "tcp",
	Addr:   "127.0.0.1:3306",
	DBName: "sql_wrapper_test",
}

// ---------- Tests ----------

func TestInsert(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())
}

func TestGet(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "John", Age: 25, Weather: Spring, Hidden: "abc"}

	// There should be no objects in the schema
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)
}

func TestUpdate(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "Luke", Age: 30, Weather: Winter, Hidden: "abc"}

	// There should be no objects present in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert one object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Change the object
	obj.Name = "Nathan"
	obj.Age = 100
	obj.Weather = Autumn

	// Update the object
	err = wrapper.Update(&obj)
	assert.Nil(err)

	// Test getting the updated object
	rows, err = database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Scan in the element
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)
}

func TestDelete(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "Steve", Age: 50, Weather: Summer, Hidden: "abc"}

	// There should be zero objects in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object present in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)

	// Delete the object from the wrapper
	err = wrapper.Delete(&obj)
	assert.Nil(err)

	// There should be zero objects in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Test the SQL to make sure there are no entries
	rows, err = database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	assert.False(rows.Next())
	assert.Nil(rows.Err())
}

func TestRead(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "Steve", Age: 50, Weather: Summer, Hidden: "abc"}

	// There should be zero objects in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object present in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)

	// Create a new wrapper
	wrapper2, err := sql_wrapper.NewWrapper[*TestObject](database, TestObject{})
	assert.Nil(err)

	// Read in using the second wrapper
	assert.Nil(wrapper2.Read())

	// There should be one object present in the new wrapper
	objs, err = wrapper2.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)
}

func TestSave(t *testing.T) {
	setup()
	assert := assert.New(t)

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}

	// Save the test object
	err := wrapper.Save(&obj)
	assert.Nil(err)

	// Get the ID of the object
	objID, err := wrapper.GetID(&obj)
	assert.Nil(err)

	// Test that the SQL database has the right entries
	rows, err := database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id      int
		name    string
		age     int
		weather Season
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)

	// Change the test object
	obj.Name = "John"
	obj.Age = 70
	obj.Weather = Spring

	err = wrapper.Save(&obj)
	assert.Nil(err)

	// Test that the SQL database has the updated entries
	rows, err = database.Query("SELECT * FROM TestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &name, &age, &weather))

	// Expect that it's equal to the object
	assert.Equal(objID, id)
	assert.Equal(obj.Name, name)
	assert.Equal(obj.Age, age)
	assert.Equal(obj.Weather, weather)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test getting the updated object
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.NotNil(objs)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)
}

// ---------- Test Setup ----------

func setup() {
	// Begin a transaction
	tx, err := database.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Drop the current wrapper
	_, err = database.Exec("DROP TABLE IF EXISTS ReferenceObject;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec("DROP TABLE IF EXISTS ReferenceObjectTestObject;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec("DROP TABLE IF EXISTS TestObject;")
	if err != nil {
		log.Fatal(err)
	}

	// Rollback the transcation on a panic
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()

	// Create the new wrapper
	wrapper, err = sql_wrapper.NewWrapper[*TestObject](database, TestObject{})
	if err != nil {
		log.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	// Connect to a testing database
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	database = db
	defer database.Close()

	// Run the tests
	m.Run()
}
