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
1 - Get posts
2 - Add post
3 - Update post
4 - Delete post
5 - View Users
6 - Add User
7 - Delete User
8 - Add Post to User
9 - Remove Post from User
`

const TypePrompt = `Enter post Type:
1 - Original
2 - Comment
3 - Repost
`

// ---------- Types ----------

type PostType string

const (
	Undefined PostType = "Undefined"
	Original  PostType = "Original"
	Comment   PostType = "Comment"
	Repost    PostType = "Repost"
)

// Post type represents a post that fits into the SQL database
type Post struct {
	Message string   `sql:"Message" def:"VARCHAR(128)"`
	Likes   int      `sql:"Likes" def:"INT"`
	Type    PostType `sql:"Type" def:"ENUM('Original', 'Comment', 'Repost')"`

	Temporary string `sql:"-"`
}

func (r Post) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM Post")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id      int
		message string
		likes   int
		t       PostType
	)
	for rows.Next() {
		if err := rows.Scan(&id, &message, &likes, &t); err != nil {
			return items, err
		}

		obj := Post{Message: message, Likes: likes, Type: t}
		items[id] = &obj
	}

	return items, nil
}

// User type represents a user in the system (one-to-many relationship)
type User struct {
	Name  string  `sql:"Name" def:"VARCHAR(128)"`
	Posts []*Post `sql:"PostID" rel:"one-to-many"`
}

func (r User) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
	rows, err := db.Query("SELECT * FROM User")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		id   int
		name string
	)
	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			return items, err
		}

		obj := User{Name: name}
		items[id] = &obj
	}

	// Query the related elements
	rows, err = db.Query("SELECT * FROM UserPost")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read for each row
	var (
		userID int
		postID int
	)
	for rows.Next() {
		// Scan in row elements
		if err := rows.Scan(&userID, &postID); err != nil {
			return items, err
		}

		// Get the refereneced objects from another wrapper
		readable, err := sql_wrapper.GetObjectBySchema("Post", postID)
		if err != nil {
			return items, err
		}
		obj, ok := readable.(*Post)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *Post")
		}

		// Get the reference we want to add this object to
		readable, ok = items[userID]
		if !ok {
			return items, fmt.Errorf("object was not saved from main table correctly (was it saved as a pointer?)")
		}
		user := readable.(*User)

		// Add the object to the corresponding reference object
		user.Posts = append(user.Posts, obj)
	}

	return items, nil
}

// ---------- Methods ----------

func getPosts() {
	// Get the posts
	posts, err := postWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving posts from SQL: %v\n", err.Error())
	}

	// Print the posts
	fmt.Println("    ID    |  Message  |   Likes   |   Type   ")
	fmt.Println("--------------------------------------------")
	for id, v := range posts {
		fmt.Printf("%-10v|%-11v|%-11v|%-10v\n", id, v.Message, v.Likes, v.Type)
	}
}

func addPost() {
	post := Post{}

	// Get the author for the post
	fmt.Print("Enter post message: ")
	scanner.Scan()

	post.Message = scanner.Text()

	// Get the number of likes
	fmt.Print("Enter amount of likes: ")
	scanner.Scan()

	likes, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	post.Likes = int(likes)

	// Get the type
	fmt.Print(TypePrompt)
	scanner.Scan()

	choice := scanner.Text()
	switch choice {
	case "1":
		post.Type = Original
	case "2":
		post.Type = Comment
	case "3":
		post.Type = Repost
	}

	// Add the post to the schema
	err = postWrapper.Save(&post)
	if err != nil {
		fmt.Printf("Error saving post: %v\n", err.Error())
	} else {
		fmt.Println("Post saved successfully")
	}
}

func updatePost() {
	// Get the post ID to update
	fmt.Print("Enter post ID to update: ")
	scanner.Scan()

	updateID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the post to update
	posts, err := postWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving posts from SQL: %v\n", err.Error())
	}

	var post *Post
	for id, v := range posts {
		if int(updateID) == id {
			post = v
			break
		}
	}

	// Ask for values to update
	fmt.Print("Enter new post message: ")
	scanner.Scan()

	post.Message = scanner.Text()

	// Get the number of likes
	fmt.Print("Enter new amount of likes: ")
	scanner.Scan()

	likes, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}
	post.Likes = int(likes)

	// Get the type
	fmt.Print(TypePrompt)
	scanner.Scan()

	choice := scanner.Text()
	switch choice {
	case "1":
		post.Type = Original
	case "2":
		post.Type = Comment
	case "3":
		post.Type = Repost
	}

	// Update the post
	err = postWrapper.Save(post)
	if err != nil {
		fmt.Printf("Error updating post: %v\n", err.Error())
	} else {
		fmt.Println("post saved successfully")
	}
}

func deletePost() {
	// Get the post ID to delete
	fmt.Print("Enter post ID to delete: ")
	scanner.Scan()

	deleteID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Find the post to delete
	posts, err := postWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving posts from SQL: %v\n", err.Error())
	}

	var post *Post
	for id, v := range posts {
		if int(deleteID) == id {
			post = v
			break
		}
	}

	// Delete the post
	if err = postWrapper.Delete(post); err != nil {
		fmt.Printf("Error deleting post: %v\n", err.Error())
	} else {
		fmt.Println("post deleted successfully")
	}
}

func getUsers() {
	// Get the users
	users, err := userWrapper.Get()
	if err != nil {
		fmt.Printf("Error in receiving users from SQL: %v\n", err.Error())
	}

	// Print the users
	fmt.Println("    ID    |   Name   |  Post IDs  ")
	fmt.Println("----------------------------------")
	for id, v := range users {
		postIDs := ""
		for _, p := range v.Posts {
			id, err := postWrapper.GetID(p)
			if err != nil {
				fmt.Printf("Error getting post ID: %v\n", err.Error())
				return
			}
			postIDs += fmt.Sprint(id) + " "
		}
		fmt.Printf("%-10v|%-10v|%v\n", id, v.Name, postIDs)
	}
}

func addUser() {
	user := User{}

	// Get the username
	fmt.Print("Enter username: ")
	scanner.Scan()

	user.Name = scanner.Text()

	// Add the user to the schema
	if err := userWrapper.Save(&user); err != nil {
		fmt.Printf("Error saving user: %v\n", err.Error())
	} else {
		fmt.Println("User saved successfully")
	}
}

func deleteUser() {
	// Get the user ID to delete
	fmt.Print("Enter user ID to delete: ")
	scanner.Scan()

	deleteID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Get the user to delete
	user, err := userWrapper.GetByID(int(deleteID))
	if err != nil {
		fmt.Printf("Error getting user with ID: %v\n", err.Error())
		return
	}

	if err = userWrapper.Delete(user); err != nil {
		fmt.Printf("Error deleting user: %v\n", err.Error())
	} else {
		fmt.Println("User deleted successfully")
	}
}

func addPostToUser() {
	// Get the user id
	fmt.Print("Enter user ID to link post to: ")
	scanner.Scan()

	userID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Get the user to link
	user, err := userWrapper.GetByID(int(userID))
	if err != nil {
		fmt.Printf("Error getting user with ID: %v\n", err.Error())
		return
	}

	// Get the post ID
	fmt.Print("Enter post ID to add: ")
	scanner.Scan()

	postID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Get the post to link
	post, err := postWrapper.GetByID(int(postID))
	if err != nil {
		fmt.Printf("Error getting post with ID: %v\n", err.Error())
		return
	}

	// Add the post to the user
	user.Posts = append(user.Posts, post)
	if err = userWrapper.Save(user); err != nil {
		fmt.Printf("Error saving post to user: %v\n", err.Error())
	} else {
		fmt.Println("Post saved successfully")
	}
}

func removePostFromUser() {
	// Get the user id
	fmt.Print("Enter user ID to link post to: ")
	scanner.Scan()

	userID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Get the user to remove the post from
	user, err := userWrapper.GetByID(int(userID))
	if err != nil {
		fmt.Printf("Error getting user with ID: %v\n", err.Error())
		return
	}

	// Get the post ID to remove
	fmt.Print("Enter post ID to add: ")
	scanner.Scan()

	postID, err := strconv.ParseInt(scanner.Text(), 10, 0)
	if err != nil {
		fmt.Println("Error reading input, please try again")
		return
	}

	// Get the post to link
	post, err := postWrapper.GetByID(int(postID))
	if err != nil {
		fmt.Printf("Error getting post with ID: %v\n", err.Error())
		return
	}

	// Remove the post to the user
	for i, p := range user.Posts {
		if p == post {
			user.Posts = append(user.Posts[:i], user.Posts[i+1:]...)
			break
		}
	}

	// Save the user
	if err = userWrapper.Save(user); err != nil {
		fmt.Printf("Error saving post to user: %v\n", err.Error())
	} else {
		fmt.Println("Post saved successfully")
	}
}

// ---------- Globals ----------
var scanner *bufio.Scanner
var postWrapper *sql_wrapper.Wrapper[*Post]
var userWrapper *sql_wrapper.Wrapper[*User]

// ---------- main -----------

func main() {
	// Open SQL database
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	scanner = bufio.NewScanner(os.Stdin)

	// Create new wrappers
	postWrapper, err = sql_wrapper.NewWrapper[*Post](db, Post{})
	if err != nil {
		log.Fatal(err)
	}

	userWrapper, err = sql_wrapper.NewWrapper[*User](db, User{})
	if err != nil {
		log.Fatal(err)
	}

	// Read in information from the wrappers
	if err := postWrapper.Read(); err != nil {
		log.Fatal(err)
	}
	if err := userWrapper.Read(); err != nil {
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
			getPosts()

		case "2":
			addPost()

		case "3":
			updatePost()

		case "4":
			deletePost()

		case "5":
			getUsers()

		case "6":
			addUser()

		case "7":
			deleteUser()

		case "8":
			addPostToUser()

		case "9":
			removePostFromUser()

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
