package sql_wrapper_test

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	sql_wrapper "github.com/ethanbaker/sql-wrapper"
	"github.com/stretchr/testify/assert"
)

// ---------- Types ----------

// ReferenceObject is used to test foreign relations
type ReferenceObject struct {
	OneToOne  *TestObject   `sql:"OneToOneID" rel:"one-to-one"`
	ManyToOne *TestObject   `sql:"ManyToOneID" rel:"many-to-one"`
	OneToMany []*TestObject `sql:"OneToManyID" rel:"one-to-many"`
	// ManyToMany
}

// Read reads in TestObjects from an SQL query
func (t ReferenceObject) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM ReferenceObject")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id          int
		oneToOneID  int
		manyToOneID int
	)
	for rows.Next() {
		// Scan in row elements
		if err := rows.Scan(&id, &oneToOneID, &manyToOneID); err != nil {
			return items, err
		}

		// Get the refereneced objects from another schema
		readable, err := sql_wrapper.GetObjectBySchema("TestObject", oneToOneID)
		if err != nil {
			return items, err
		}
		oneToOne, ok := readable.(*TestObject)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *TestObject")
		}

		readable, err = sql_wrapper.GetObjectBySchema("TestObject", manyToOneID)
		if err != nil {
			return items, err
		}
		manyToOne, ok := readable.(*TestObject)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *TestObject")
		}

		// Create and add the reference object
		obj := ReferenceObject{OneToOne: oneToOne, ManyToOne: manyToOne}
		obj.OneToMany = make([]*TestObject, 0)
		items[id] = &obj
	}

	// Query the related elements
	rows, err = db.Query("SELECT * FROM ReferenceObjectTestObject")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		referenceObjectID int
		testObjectID      int
	)
	for rows.Next() {
		// Scan in row elements
		if err := rows.Scan(&referenceObjectID, &testObjectID); err != nil {
			return items, err
		}

		// Get the refereneced objects from another schema
		readable, err := sql_wrapper.GetObjectBySchema("TestObject", testObjectID)
		if err != nil {
			return items, err
		}
		obj, ok := readable.(*TestObject)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *TestObject")
		}

		// Get the reference we want to add this object to
		readable, ok = items[referenceObjectID]
		if !ok {
			return items, fmt.Errorf("referenceObjectID is not in main table")
		}
		ref := readable.(*ReferenceObject)

		// Add the object to the corresponding reference object
		ref.OneToMany = append(ref.OneToMany, obj)
	}

	return items, nil
}

// ---------- Globals ----------

var referenceWrapper *sql_wrapper.Wrapper[*ReferenceObject]

// ---------- Tests ----------

func TestInsertWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj := TestObject{Name: "Jack", Age: 20, Weather: Summer, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj, ManyToOne: &obj, OneToMany: []*TestObject{&obj}}

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Insert the reference object
	refID, err := referenceWrapper.Insert(&ref)
	assert.Nil(err)

	// Test that the SQL TestObject database has the right entries
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

	// Test that the SQL ReferenceObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		oneToOneID  int
		manyToOneID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the object
	assert.Equal(refID, id)
	assert.Equal(objID, oneToOneID)
	assert.Equal(objID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		referenceObjectID int
		testObjectID      int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to the object
	assert.Equal(refID, referenceObjectID)
	assert.Equal(objID, testObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())
}

func TestGetWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj := TestObject{Name: "John", Age: 25, Weather: Spring, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj, ManyToOne: &obj, OneToMany: []*TestObject{&obj}}

	// There should be no objects in either wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	refObjs, err := referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(refObjs))

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Insert the reference object
	refID, err := referenceWrapper.Insert(&ref)
	assert.Nil(err)

	// There should be one object in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(obj.Name, objs[objID].Name)
	assert.Equal(obj.Age, objs[objID].Age)
	assert.Equal(obj.Weather, objs[objID].Weather)

	// There should be one object in the reference wrapper
	refObjs, err = referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(refObjs))
	assert.Equal(ref.OneToOne, refObjs[refID].OneToOne)
	assert.Equal(ref.ManyToOne, refObjs[refID].ManyToOne)
	assert.Equal(len(ref.OneToMany), len(refObjs[refID].OneToMany))
	assert.True(len(ref.OneToMany) == 1 && len(refObjs[refID].OneToMany) == 1)
	assert.Equal(ref.OneToMany[0], refObjs[refID].OneToMany[0])
}

func TestUpdateWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj1 := TestObject{Name: "Luke", Age: 30, Weather: Winter, Hidden: "abc"}
	obj2 := TestObject{Name: "John", Age: 10, Weather: Summer, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj1, ManyToOne: &obj1, OneToMany: []*TestObject{&obj1}}

	// There should be no objects present in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert two objects
	obj1ID, err := wrapper.Insert(&obj1)
	assert.Nil(err)

	obj2ID, err := wrapper.Insert(&obj2)
	assert.Nil(err)

	// Insert the reference object
	refID, err := referenceWrapper.Insert(&ref)
	assert.Nil(err)

	// Test that the SQL ReferenceObject database has the right entries
	rows, err := database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id          int
		oneToOneID  int
		manyToOneID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the reference object
	assert.Equal(refID, id)
	assert.Equal(obj1ID, oneToOneID)
	assert.Equal(obj1ID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		testObjectID      int
		referenceObjectID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 1
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj1ID, referenceObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(2, len(objs))
	assert.Equal(obj1.Name, objs[obj1ID].Name)
	assert.Equal(obj1.Age, objs[obj1ID].Age)
	assert.Equal(obj1.Weather, objs[obj1ID].Weather)
	assert.Equal(obj2.Name, objs[obj2ID].Name)
	assert.Equal(obj2.Age, objs[obj2ID].Age)
	assert.Equal(obj2.Weather, objs[obj2ID].Weather)

	// There should be one object in the reference wrapper
	refObjs, err := referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(refObjs))
	assert.Equal(ref.OneToOne, refObjs[refID].OneToOne)
	assert.Equal(ref.ManyToOne, refObjs[refID].ManyToOne)
	assert.Equal(len(ref.OneToMany), len(refObjs[refID].OneToMany))
	assert.True(len(ref.OneToMany) == 1 && len(refObjs[refID].OneToMany) == 1)
	assert.Equal(ref.OneToMany[0], refObjs[refID].OneToMany[0])

	// Change the reference object
	ref.OneToOne = &obj2
	ref.ManyToOne = &obj2
	ref.OneToMany = append(ref.OneToMany, &obj2)

	// Update the object
	err = referenceWrapper.Update(&ref)
	assert.Nil(err)

	// Test getting the updated object
	rows, err = database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the reference object
	assert.Equal(refID, id)
	assert.Equal(obj2ID, oneToOneID)
	assert.Equal(obj2ID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has been updated
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read the first entry
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 1
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj1ID, testObjectID)

	// Read in the second entry
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 2
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj2ID, testObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Make sure the wrapper has the updated value
	refObjs, err = referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(refObjs))
	assert.Equal(ref.OneToOne, refObjs[refID].OneToOne)
	assert.Equal(ref.ManyToOne, refObjs[refID].ManyToOne)
	assert.Equal(len(ref.OneToMany), len(refObjs[refID].OneToMany))
	assert.True(len(ref.OneToMany) == 2 && len(refObjs[refID].OneToMany) == 2)
	assert.Equal(ref.OneToMany[0], refObjs[refID].OneToMany[0])
	assert.Equal(ref.OneToMany[1], refObjs[refID].OneToMany[1])
}

func TestDeleteWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj := TestObject{Name: "Steve", Age: 50, Weather: Summer, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj, ManyToOne: &obj, OneToMany: []*TestObject{&obj}}

	// There should be zero objects in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Insert the reference object
	_, err = referenceWrapper.Insert(&ref)
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

	// Cannot delete an object while a foreign object references it
	err = wrapper.Delete(&obj)
	assert.NotNil(err)

	// Delete the reference object
	err = referenceWrapper.Delete(&ref)
	assert.Nil(err)

	// There should be zero objects in the wrapper
	refObjs, err := referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(refObjs))

	// Test the SQL ReferenceObject table to make sure there are no entries
	rows, err = database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test the SQL ReferenceObjectTestObject table to make sure there are no entries
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	assert.False(rows.Next())
	assert.Nil(rows.Err())
}

func TestReadWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj := TestObject{Name: "Steve", Age: 50, Weather: Summer, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj, ManyToOne: &obj, OneToMany: []*TestObject{&obj}}

	// Insert the test object
	objID, err := wrapper.Insert(&obj)
	assert.Nil(err)

	// Insert the reference object
	refID, err := referenceWrapper.Insert(&ref)
	assert.Nil(err)

	// Test that the SQL TestObject database has the right entries
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

	// Test that the SQL ReferenceObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		oneToOneID  int
		manyToOneID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the object
	assert.Equal(refID, id)
	assert.Equal(objID, oneToOneID)
	assert.Equal(objID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		referenceObjectID int
		testObjectID      int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to the object
	assert.Equal(refID, referenceObjectID)
	assert.Equal(objID, testObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Create a new wrapper
	newWrapper, err := sql_wrapper.NewWrapper[*ReferenceObject](database, ReferenceObject{})
	assert.Nil(err)

	// Read in using the second wrapper
	assert.Nil(newWrapper.Read())

	// There should be one object present in the new wrapper
	objs, err := newWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(objs))
	assert.Equal(&obj, objs[refID].OneToOne)
	assert.Equal(&obj, objs[refID].ManyToOne)
	assert.Equal(1, len(objs[refID].OneToMany))
	assert.Equal(&obj, objs[refID].OneToMany[0])
}

func TestSaveWithForeignRelation(t *testing.T) {
	referenceSetup()
	assert := assert.New(t)

	obj1 := TestObject{Name: "Luke", Age: 30, Weather: Winter, Hidden: "abc"}
	obj2 := TestObject{Name: "John", Age: 10, Weather: Summer, Hidden: "abc"}
	ref := ReferenceObject{OneToOne: &obj1, ManyToOne: &obj1, OneToMany: []*TestObject{&obj1}}

	// There should be no objects present in the wrapper
	objs, err := wrapper.Get()
	assert.Nil(err)
	assert.Equal(0, len(objs))

	// Save two objects
	err = wrapper.Save(&obj1)
	assert.Nil(err)

	err = wrapper.Save(&obj2)
	assert.Nil(err)

	// Save the reference object
	err = referenceWrapper.Save(&ref)
	assert.Nil(err)

	// Get the IDs of the objects for testing
	obj1ID, err := wrapper.GetID(&obj1)
	assert.Nil(err)

	obj2ID, err := wrapper.GetID(&obj2)
	assert.Nil(err)

	refID, err := referenceWrapper.GetID(&ref)
	assert.Nil(err)

	// Test that the SQL ReferenceObject database has the right entries
	rows, err := database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		id          int
		oneToOneID  int
		manyToOneID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the reference object
	assert.Equal(refID, id)
	assert.Equal(obj1ID, oneToOneID)
	assert.Equal(obj1ID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has the right entries
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read in the first entry
	var (
		testObjectID      int
		referenceObjectID int
	)
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 1
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj1ID, referenceObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// There should be one object in the wrapper
	objs, err = wrapper.Get()
	assert.Nil(err)
	assert.Equal(2, len(objs))
	assert.Equal(obj1.Name, objs[obj1ID].Name)
	assert.Equal(obj1.Age, objs[obj1ID].Age)
	assert.Equal(obj1.Weather, objs[obj1ID].Weather)
	assert.Equal(obj2.Name, objs[obj2ID].Name)
	assert.Equal(obj2.Age, objs[obj2ID].Age)
	assert.Equal(obj2.Weather, objs[obj2ID].Weather)

	// There should be one object in the reference wrapper
	refObjs, err := referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(refObjs))
	assert.Equal(ref.OneToOne, refObjs[refID].OneToOne)
	assert.Equal(ref.ManyToOne, refObjs[refID].ManyToOne)
	assert.Equal(len(ref.OneToMany), len(refObjs[refID].OneToMany))
	assert.True(len(ref.OneToMany) == 1 && len(refObjs[refID].OneToMany) == 1)
	assert.Equal(ref.OneToMany[0], refObjs[refID].OneToMany[0])

	// Change the reference object
	ref.OneToOne = &obj2
	ref.ManyToOne = &obj2
	ref.OneToMany = append(ref.OneToMany, &obj2)

	// Update the object
	err = referenceWrapper.Save(&ref)
	assert.Nil(err)

	// Test getting the updated object
	rows, err = database.Query("SELECT * FROM ReferenceObject")
	assert.Nil(err)
	defer rows.Close()

	assert.True(rows.Next())
	assert.Nil(rows.Scan(&id, &oneToOneID, &manyToOneID))

	// Expect that it's equal to the reference object
	assert.Equal(refID, id)
	assert.Equal(obj2ID, oneToOneID)
	assert.Equal(obj2ID, manyToOneID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Test that the SQL ReferenceObjectTestObject database has been updated
	rows, err = database.Query("SELECT * FROM ReferenceObjectTestObject")
	assert.Nil(err)
	defer rows.Close()

	// Read the first entry
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 1
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj1ID, testObjectID)

	// Read in the second entry
	assert.True(rows.Next())
	assert.Nil(rows.Scan(&referenceObjectID, &testObjectID))

	// Expect that it's equal to object 2
	assert.Equal(refID, referenceObjectID)
	assert.Equal(obj2ID, testObjectID)

	// There should be no more objects in the query
	assert.False(rows.Next())
	assert.Nil(rows.Err())

	// Make sure the wrapper has the updated value
	refObjs, err = referenceWrapper.Get()
	assert.Nil(err)
	assert.Equal(1, len(refObjs))
	assert.Equal(ref.OneToOne, refObjs[refID].OneToOne)
	assert.Equal(ref.ManyToOne, refObjs[refID].ManyToOne)
	assert.Equal(len(ref.OneToMany), len(refObjs[refID].OneToMany))
	assert.True(len(ref.OneToMany) == 2 && len(refObjs[refID].OneToMany) == 2)
	assert.Equal(ref.OneToMany[0], refObjs[refID].OneToMany[0])
}

// ---------- Test Setup ----------

func referenceSetup() {
	// Begin a transaction
	tx, err := database.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Drop the current wrapper
	_, err = database.Exec("DROP TABLE IF EXISTS ReferenceObjectTestObject;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec("DROP TABLE IF EXISTS ReferenceObject;")
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

	// Create the new wrappers
	wrapper, err = sql_wrapper.NewWrapper[*TestObject](database, TestObject{})
	if err != nil {
		log.Fatal(err)
	}

	referenceWrapper, err = sql_wrapper.NewWrapper[*ReferenceObject](database, ReferenceObject{})
	if err != nil {
		log.Fatal(err)
	}
}
