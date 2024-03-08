<!--
  Created by: Ethan Baker (contact@ethanbaker.dev)
  
  Adapted from:
    https://github.com/othneildrew/Best-README-Template/

Here are different preset "variables" that you can search and replace in this template.
`path_to_logo`
`path_to_demo`
-->

<div id="top"></div>

<!-- PROJECT SHIELDS/BUTTONS -->
<!-- 
  Netlify buttons:
[![Netlify Status]()]()
  Golang specific buttons:
-->
![0.1.0](https://img.shields.io/badge/status-0.1.0-red)
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]


[![GoDoc](https://godoc.org/github.com/ethanbaker/sql-wrapper?status.svg)](https://godoc.org/github.com/ethanbaker/sql-wrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/sql-wrapper)](https://goreportcard.com/report/github.com/ethanbaker/sql-wrapper)
[![Go Coverage](./docs/go-coverage.svg)](./docs/go-coverage.svg)


<!-- PROJECT LOGO -->
<br><br><br>
<div align="center">
  <a href="https://github.com/ethanbaker/sql-wrapper">
    <img src="./docs/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">SQL Wrapper</h3>

  <p align="center">
    An SQL wrapper for structs in Golang
  </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About</a>
      <ul>
        <li><a href="#limitations">Limitations</a></li>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#struct-setup">Struct Setup</a></li>
        <li><a href="#foreign-relations">Foreign Relations</a></li>
        <li><a href="#read-method">Read Method</a></li>
        <li><a href="#creating-new-wrappers">Creating New Wrappers</a></li>
        <li><a href="#wrapper-functions">Wrapper Functions</a></li>
        <li><a href="#examples">Examples</a></li>
      </ul>
    </li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>


<!-- ABOUT -->
## About

This project is used to wrap golang structs in a schema that makes it easier to repeatedly save occurances of those structs to SQL. 

The motivation of this project is to avoid the annoying task of updating SQL saving in golang projects. For example, if an API is saving `Record` structs to an SQL database, and then spec changes require the `Record` struct to change, it can be very cumbersome to go through all of the manually-written SQL code to change this.

In addition, wrapping SQL structs makes set up much, much easier. You only need to worry about the initialization of your structs and related SQL instead of every SQL action.

<p align="right">(<a href="#top">back to top</a>)</p>


### Limitations

This project has a few key limitations:
* You must define all of your SQL types in the struct using tags
  * The wrapper attempts to match these automatically, but it trusts SQL to make decisions and throw errors. If you declare a field as an integer when it is actually a string, SQL will handle it
* You must make your struct a part of the `Readable` interface
  * This is required because I have no idea how to automate reading SQL tables with an indeterminate amount of interface pointers (if you know how to do this please consider contributing!)
  * There are some "template" `Read` methods in the `examples` directory for different scenarios that you can check out
* Not every struct attribute is supported
  * So far, only primitive types (`string`, `int`, `enums`, etc), pointers (`*Object`), and lists of pointers (`[]*Object`) are supported
  * Maps are **not** supported

<p align="right">(<a href="#top">back to top</a>)</p>

### Built With

* [Golang](https://go.dev/)
* SQL

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

This wrapper is relatively easy to set up for your projects. These steps will follow along the `examples/record` project found in the repository.

#### Struct Setup

For each struct you want to wrap, define the struct and the SQL table it represents. At the moment this is a relatively manual process done with struct tags.
* The `sql` tag tells the wrapper to include this field for SQL consideration. The value of this field is the name of the associated SQL column
* The `def` tag tells the wrapper how to initialize this field as a column in SQL.

You do not need to create an ID field; one will be added automatically.

```go
type Record struct {
  Author string   `sql:"Author" def:"VARCHAR(128)"`
	Likes  int      `sql:"Likes" def:"INT"`
	Type   PostType `sql:"Type" def:"ENUM('Original', 'Comment', 'Repost')"`
}
```

In this instance, a `Record` struct is created with three different fields (the `Type` field has an enum value defined outside of this snippet). Then, each field has related tags to fit with the definition.

When the wrapper is initialized, an SQL table will be created that looks like the following:

|id |Author|Likes|Type|
|---|------|-----|----|
|...|      |     |    |

Each column has the following SQL type:
* **id**: `INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY` (this is automatically generated for every struct)
* **Author**: `VARCHAR(128)`
* **Likes**: `INT`
* **Type**: `ENUM('Original', 'Comment', 'Repost')`

#### Foreign Relations

Foreign relations are supported for wrappers. Foreign relations work by relating a **source** (the struct you are defining) and a **target** (the struct you are referencing).

There are four different types stored as the `RelationType` enum:
* `OneToOne`: Match one source object to one target object. Objects cannot be linked to more than one other object.
  * Relations are done using a *pointer* to the other object
  * This is equivalent to `ManyToOne` with a `UNIQUE` constraint on the target ID
* `ManyToOne`: Match any amount of source objects to one target object. Source objects cannot be linked to more than one other object.
  * Relations are done using a *pointer* to the other object
* `OneToMany`: Match any amount of target objects to one source object. Target objects cannot be linked to more than one source object.
  * Relations are done using an *array of pointers* to the other object(s)
  * This is equivalent to `ManyToMany` with a `UNIQUE` constraint on the target ID
* `ManyToMany`: Match any amount of target objects to any source object.
  * Relations are done using an *array of pointers* to the other object(s)

You can define struct attributes to be a foreign relation by adding a `rel` field to the attribute tag. For example:

```go
type ReferenceObject struct {
	OneToOne   *Object1   `sql:"Object1ID" rel:"one-to-one"`
	ManyToOne  *Object2   `sql:"Object2ID" rel:"many-to-one"`
	OneToMany  []*Object3 `sql:"Object3ID" rel:"one-to-many"`
	ManyToMany []*Object4 `sql:"Object4ID" rel:"many-to-many"`
}
```

You can see examples of foreign relations in the `examples` folder of this project. The examples that deal with foreign relations are:
* `examples/user-post`: a one-to-many relationship between User and Post, where a User can have a list of Posts
* `examples/item-identification`: a one-to-one relationship between Item and Identification. An Item has one and only one Identification struct created and linked to it

#### Read Method

Next, a `Read` function needs to be created that is attached to the struct.

Because SQL querires need specificity when reading in new values, this is done easiest through a user-defined function. 

If you know of a way to make reading in SQL tables easier, please consider [contributing](#contributing)!

```go
// Read function reads in values from an SQL database
// NOTE: there is no pointer receiver in this method to properly match the Readable interface
func (r Record) Read(db *sql.DB) (map[int]sql_wrapper.Readable, error) {
  // Create a list of Readable objects to populate
	items := map[int]sql_wrapper.Readable{}

	// Get the main elements
  // NOTE: the name of the SQl table you want to use is the name of the struct
  // NOTE: the commented line below should be equivalent:
  // rows, err := db.Query("SELECT * FROM " + reflect.TypeOf(s.template).Name())
	rows, err := db.Query("SELECT * FROM Record")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	// Read each row in the query
	var (
		id     int
		author string
		likes  int
		t      PostType
	)
	for rows.Next() {
    // Scan the specific elements of this struct in to custom variables
		if err := rows.Scan(&id, &author, &likes, &t); err != nil {
			return items, err
		}

    // Create a new Record object and add it to our list
    // NOTE: you **MUST** add the object as a pointer. If not, then the object will be stored and retreived as a copy, leading to undesired behavior
		obj := Record{Author: author, Likes: likes, Type: t}
		items[id] = &obj
	}

	return items, nil
}
```

As noted in the code snippet, there are many important notes to keep in mind:
* The function is defined as `func (o Object) Read...`. It is **not** constructed with a pointer receiver (i.e. no `(o *Object)`). This is to make `Object` fit the `Readable` interface
* The SQL table queried from is the same as the object name. You can replace this with `reflect` to automate the name-getting
* You **MUST** add objects as pointers or else undesired behavior will enter your program and it will most likely not work

If you are creating a `Read` method with foreign relations, you need to perform more operations than a standard read.

For pointer relationships (**one-to-one** and **many-to-one**), you must get the referenced object by using the public method `GetObjectBySchema` and then cast it to the object you want. Here is an example from `examples/item-identification`:

```go
// Read reads in SQL values to the wrapper
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

		// Get the referenced item. We use a public method together with the name of the database/wrapper we want to find the link to. This gives us an object of the Readable interface
		readable, err := sql_wrapper.GetObjectBySchema("Item", itemID)
		if err != nil {
			return items, err
		}

    // Now we can cast our readable object to the pointer we want
    // NOTE: this cast will fail if you didn't save your object as a pointer!
		item, ok := readable.(*Item)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *post")
		}

		// Create and add the object
    // NOTE: save the object as a pointer!
		obj := Identification{Number: number, Hash: hash, Item: item}
		items[id] = &obj
	}

	return items, nil
}
```

For slice relationships (**one-to-many** and **many-to-many**), you must get the rows of another SQL table that links the two wrappers together. This table is defined as the source struct's name concatenated with the target's struct name. Here is an example from `examples/user-post`:

```go
// Read reads in SQL values to the wrapper
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

    // Add the object normally
		obj := User{Name: name}
		items[id] = &obj
	}

	// Query the related elements
  // NOTE: the source, or the struct we're writing Read for, is User. The target, or the struct being referenced by the source, is Post. So, the table name is UserPost, and ID columns follow the same order
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

    // Get the reference object from the other wrapper
		readable, err := sql_wrapper.GetObjectBySchema("Post", postID)
		if err != nil {
			return items, err
		}

    // Cast it to be the object we want
		obj, ok := readable.(*Post)
		if !ok {
			return items, fmt.Errorf("cannot cast object to *Post")
		}

    // Get the user we want to link this item to
		readable, ok = items[userID]
		if !ok {
			return items, fmt.Errorf("object was not saved from main table correctly (was it saved as a pointer?)")
		}
		user := readable.(*User)

		// Add the target object to the corresponding source object
		user.Posts = append(user.Posts, obj)
	}

	return items, nil
}
```

#### Creating New Wrappers

Now that your struct has been finished, a wrapper must be created to keep your struct objects synced to SQL. Creating a wrapper can be done as so with the required parameters.

```go
wrapper, err := sql_wrapper.NewWrapper[*Record](db, Record{})
```

In order to create a schema, you must provide:
* The SQL database needed to execute statements
* An empty struct you are parametrizing the schema with
* A generic type as a pointer, which lets the wrapper know what type to return

You can read in existing SQL entries using the `Read` function:

```go
if err := wrapper.Read(); err != nil {
  // Handle error
}
```

You should call `Read` on wrappers without foreign references **first**. This allows other wrappers with foreign references to pull in relations after the other wrapper has loaded first.

#### Wrapper Functions

After your wrapper is created, you can then call functions associated with it. These are present in the examples and the [documentation][documentation-url].

<p align="right">(<a href="#top">back to top</a>)</p>

### Examples

There are numerous examples present in the `examples` directory. You can check these out for help with your own project. You can do this by editing the SQL config struct present in the examples with your own values initialized on your own machine.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [x] Foreign Key Constraints
- [x] Find by ID Functions
- [ ] Foreign Relations with Maps
- [ ] Create GitHub Actions Workflow

See the [open issues][issues-url] for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

For issues and suggestions, please include as much useful information as possible.
Review the [documentation][documentation-url] and make sure the issue is actually
present or the suggestion is not included. Please share issues/suggestions on the
[issue tracker][issues-url].

For patches and feature additions, please submit them as [pull requests][pulls-url]. 
Please adhere to the [conventional commits][conventional-commits-url]. standard for
commit messaging. In addition, please try to name your git branch according to your
new patch. [These standards][conventional-branches-url] are a great guide you can follow.

You can follow these steps below to create a pull request:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b branch_name`)
3. Commit your Changes (`git commit -m "commit_message"`)
4. Push to the Branch (`git push origin branch_name`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- LICENSE -->
## License

This project uses the Apache 2.0 License.

You can find more information in the [LICENSE][license-url] file.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Ethan Baker - contact@ethanbaker.dev - [LinkedIn][linkedin-url]

Project Link: [https://github.com/ethanbaker/sql-wrapper][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/sql-wrapper.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/sql-wrapper.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/sql-wrapper.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/sql-wrapper.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/sql-wrapper.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: <https://github.com/ethanbaker/sql-wrapper/graphs/contributors>
[forks-url]: <https://github.com/ethanbaker/sql-wrapper/network/members>
[stars-url]: <https://github.com/ethanbaker/sql-wrapper/stargazers>
[issues-url]: <https://github.com/ethanbaker/sql-wrapper/issues>
[pulls-url]: <https://github.com/ethanbaker/sql-wrapper/pulls>
[license-url]: <https://github.com/ethanbaker/sql-wrapper/blob/master/LICENSE>
[linkedin-url]: <https://linkedin.com/in/ethandbaker>
[project-url]: <https://github.com/ethanbaker/sql-wrapper>

[documentation-url]: <https://godoc.org/github.com/ethanbaker/sql-wrapper>

[conventional-commits-url]: <https://www.conventionalcommits.org/en/v1.0.0/#summary>
[conventional-branches-url]: <https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops>