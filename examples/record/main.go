package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	sql_wrapper "github.com/ethanbaker/sql-wrapper"
	"github.com/go-sql-driver/mysql"
)

// SQL credentials
var cfg = mysql.Config{
	User:   "sql_wrapper_example",
	Passwd: "abc123",
	Net:    "tcp",
	Addr:   "127.0.0.1:3306",
	DBName: "sql_wrapper_example",
}

// User Prompts
const AskPrompt = `Enter Choice:
1 - Get Records
2 - Add Record
3 - Update Record
4 - Delete Record
`

const TypePrompt = `Enter Record Type:
1 - Original
2 - Comment
3 - Repost
`

// PostType is used to encode an enum into SQL
type PostType string

const (
	Undefined PostType = "Undefined"
	Original  PostType = "Original"
	Comment   PostType = "Comment"
	Repost    PostType = "Repost"
)

// Record type represents a record that fits into the SQL database
type Record struct {
	Author string   `sql:"Author" def:"VARCHAR(128)"`
	Likes  int      `sql:"Likes" def:"INT"`
	Type   PostType `sql:"Type" def:"ENUM('Original', 'Comment', 'Repost')"`
	Hidden string   `sql:"-"`
}

func (r Record) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM Record")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id     int
		author string
		likes  int
		t      PostType
	)
	for rows.Next() {
		if err := rows.Scan(&id, &author, &likes, &t); err != nil {
			return items, err
		}

		obj := Record{Author: author, Likes: likes, Type: t}
		items[id] = &obj
	}

	return items, nil
}

func getRecords() {
	// Get the records
	records, err := wrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving records from SQL: %v\n", err.Error())
		return
	}

	// Print the records
	fmt.Println("    ID    |  Author  |   Likes   |   Type   ")
	fmt.Println("--------------------------------------------")
	for id, v := range records {
		fmt.Printf("%-10v|%-10v|%-11v|%-10v\n", id, v.Author, v.Likes, v.Type)
	}
}

func addRecord() {
	record := Record{}

	// Get the author for the record
	fmt.Print("Enter record author: ")
	scanner.Scan()

	record.Author = scanner.Text()

	// Get the number of likes
	fmt.Print("Enter amount of likes: ")
	scanner.Scan()

	likes, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	record.Likes = int(likes)

	// Get the type
	fmt.Print(TypePrompt)
	scanner.Scan()

	choice := scanner.Text()
	switch choice {
	case "1":
		record.Type = Original
	case "2":
		record.Type = Comment
	case "3":
		record.Type = Repost
	}

	// Add the record to the schema
	err = wrapper.Save(&record)
	if err != nil {
		fmt.Printf("Error saving record: %v\n", err.Error())
	} else {
		fmt.Println("Record saved successfully")
	}
}

func updateRecord() {
	// Get the record ID to update
	fmt.Print("Enter record ID to update: ")
	scanner.Scan()

	updateID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the record to update
	records, err := wrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving records from SQL: %v\n", err.Error())
	}

	var record *Record
	for id, v := range records {
		if int(updateID) == id {
			record = v
			break
		}
	}

	// Ask for values to update
	fmt.Print("Enter record author: ")
	scanner.Scan()

	record.Author = scanner.Text()

	// Get the number of likes
	fmt.Print("Enter amount of likes: ")
	scanner.Scan()

	likes, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	record.Likes = int(likes)

	// Get the type
	fmt.Print(TypePrompt)
	scanner.Scan()

	choice := scanner.Text()
	switch choice {
	case "1":
		record.Type = Original
	case "2":
		record.Type = Comment
	case "3":
		record.Type = Repost
	}

	// Update the record
	err = wrapper.Save(record)
	if err != nil {
		fmt.Printf("Error saving record: %v\n", err.Error())
	} else {
		fmt.Println("Record saved successfully")
	}
}

func deleteRecord() {
	// Get the record ID to delete
	fmt.Print("Enter record ID to delete: ")
	scanner.Scan()

	deleteID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the record to delete
	records, err := wrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving records from SQL: %v\n", err.Error())
	}

	var record *Record
	for id, v := range records {
		if int(deleteID) == id {
			record = v
			break
		}
	}

	// Delete the record
	if err = wrapper.Delete(record); err != nil {
		fmt.Printf("Error deleting record: %v\n", err.Error())
	} else {
		fmt.Println("Record deleted successfully")
	}
}

// Globals
var scanner *bufio.Scanner
var wrapper *sql_wrapper.Wrapper[*Record]

func main() {
	// Open SQL database
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(os.Stdin)
	wrapper, err = sql_wrapper.NewWrapper[*Record](db, Record{})
	if err != nil {
		log.Fatal(err)
	}

	// Read in credentials from the schema
	if err := wrapper.Read(); err != nil {
		log.Fatal(err)
	}

	// Read user input
	for {
		fmt.Print(AskPrompt)
		scanner.Scan()

		text := scanner.Text()

		// Break if user enters empty string
		if len(text) == 0 {
			break
		}

		fmt.Println()

		switch text {
		case "1":
			getRecords()

		case "2":
			addRecord()

		case "3":
			updateRecord()

		case "4":
			deleteRecord()

		default:
			continue
		}

		fmt.Println()
	}

	// handle error
	if scanner.Err() != nil {
		fmt.Println("Scanner error: ", scanner.Err())
	}
}
