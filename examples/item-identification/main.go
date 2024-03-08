package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	sql_wrapper "github.com/ethanbaker/sql-wrapper"
	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// ---------- Consts ----------

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
1 - Get Items
2 - Add Item
3 - Update Item
4 - Delete Item
5 - Get Identifications
`

// ---------- Types ----------

// Item type represents a item that fits into the SQL database
type Item struct {
	Name  string `sql:"Name" def:"VARCHAR(128)"`
	Price int    `sql:"Price" def:"INT"`

	Temporary string `sql:"-"`
}

func (r Item) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM Item")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id    int
		name  string
		price int
	)
	for rows.Next() {
		if err := rows.Scan(&id, &name, &price); err != nil {
			return items, err
		}

		obj := Item{Name: name, Price: price}
		items[id] = &obj
	}

	return items, nil
}

// Identification type represents identifying information about an item (one-to-one relationship)
type Identification struct {
	Number int    `sql:"Number" def:"INT"`
	Hash   string `sql:"Hash" def:"VARCHAR(128)"`

	Item *Item `sql:"ItemID" rel:"one-to-one"`
}

func (r Identification) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM Identification")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id     int
		number int
		hash   string
		itemID int
	)
	for rows.Next() {
		if err := rows.Scan(&id, &number, &hash, &itemID); err != nil {
			return items, err
		}

		// Get the referenced item
		readable, err := sql_wrapper.GetObjectBySchema("Item", itemID)
		if err != nil {
			return items, err
		}
		item, ok := readable.(*Item)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *post")
		}

		// Create and add the object
		obj := Identification{Number: number, Hash: hash, Item: item}
		items[id] = &obj
	}

	return items, nil
}

// ---------- Methods ----------

func getItems() {
	// Get the items
	posts, err := itemWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving items from SQL: %v\n", err.Error())
	}

	// Print the posts
	fmt.Println("    ID    |  Name  |   Price   ")
	fmt.Println("-------------------------------")
	for id, v := range posts {
		fmt.Printf("%-10v|%-10v|%v\n", id, v.Name, v.Price)
	}
}

func addItem() {
	item := Item{}

	// Get the name of the item
	fmt.Print("Enter item name: ")
	scanner.Scan()

	item.Name = scanner.Text()

	// Get the price of the item
	fmt.Print("Enter price of item: ")
	scanner.Scan()

	price, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	item.Price = int(price)

	// Add the item to the wrapper
	err = itemWrapper.Save(&item)
	if err != nil {
		fmt.Printf("Error saving item: %v\n", err.Error())
	} else {
		fmt.Println("item saved successfully")
	}

	// Add identification for the post
	iden := Identification{}
	iden.Number = rand.Intn(1000)
	iden.Hash = uuid.New().String()[:6]
	iden.Item = &item

	err = identificationWrapper.Save(&iden)
	if err != nil {
		fmt.Printf("Error saving identification: %v\n", err.Error())
	}
}

func updateItem() {
	// Get the item ID to update
	fmt.Print("Enter item ID to update: ")
	scanner.Scan()

	updateID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the item to update
	items, err := itemWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving items from SQL: %v\n", err.Error())
	}

	var item *Item
	for id, v := range items {
		if int(updateID) == id {
			item = v
			break
		}
	}

	// Ask for values to update
	fmt.Print("Enter new item name: ")
	scanner.Scan()

	item.Name = scanner.Text()

	// Get the price
	fmt.Print("Enter new price: ")
	scanner.Scan()

	price, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	item.Price = int(price)

	// Update the item
	err = itemWrapper.Save(item)
	if err != nil {
		fmt.Printf("Error saving item: %v\n", err.Error())
	} else {
		fmt.Println("Item saved successfully")
	}
}

func deleteItem() {
	// Get the item ID to delete
	fmt.Print("Enter item ID to update: ")
	scanner.Scan()

	deleteID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the post to delete
	items, err := itemWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving items from SQL: %v\n", err.Error())
	}

	var item *Item
	for id, v := range items {
		if int(deleteID) == id {
			item = v
			break
		}
	}

	// Delete the post
	if err = itemWrapper.Delete(item); err != nil {
		fmt.Printf("Error deleting post: %v\n", err.Error())
	} else {
		fmt.Println("post deleted successfully")
	}
}

func getIdentifications() {
	// Get the identifications
	identifications, err := identificationWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving identifications from SQL: %v\n", err.Error())
	}

	// Print the identifications
	fmt.Println("    ID    |  Number  |   Hash   | Item Name")
	fmt.Println("---------------------------------------------")
	for id, v := range identifications {
		fmt.Printf("%-10v|%-10v|%-10v|%-10v\n", id, v.Number, v.Hash, v.Item.Name)
	}
}

// ---------- Globals ----------
var scanner *bufio.Scanner
var itemWrapper *sql_wrapper.Wrapper[*Item]
var identificationWrapper *sql_wrapper.Wrapper[*Identification]

// ---------- main -----------

func main() {
	// Open SQL database
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(os.Stdin)

	// Create new wrappers
	itemWrapper, err = sql_wrapper.NewWrapper[*Item](db, Item{})
	if err != nil {
		log.Fatal(err)
	}

	identificationWrapper, err = sql_wrapper.NewWrapper[*Identification](db, Identification{})
	if err != nil {
		log.Fatal(err)
	}

	// Read in information from the wrappers
	if err := itemWrapper.Read(); err != nil {
		log.Fatal(err)
	}
	if err := identificationWrapper.Read(); err != nil {
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
			getItems()

		case "2":
			addItem()

		case "3":
			updateItem()

		case "4":
			deleteItem()

		case "5":
			getIdentifications()

		default:
			continue
		}

		fmt.Println()
	}

	// handle error
	if scanner.Err() != nil {
		fmt.Println("Error: ", scanner.Err())
	}
}
