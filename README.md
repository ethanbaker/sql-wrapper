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
![0.1.0](https://img.shields.io/badge/status-a.1.0-red)
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]


[![GoDoc](https://godoc.org/github.com/ethanbaker/sql_wrapper?status.svg)](https://godoc.org/github.com/ethanbaker/sql_wrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/sql_wrapper)](https://goreportcard.com/report/github.com/ethanbaker/sql_wrapper)
[![Go Coverage](./docs/go-coverage.svg)](./docs/go-coverage.svg)


<!-- PROJECT LOGO -->
<br><br><br>
<div align="center">
  <a href="https://github.com/ethanbaker/sql_wrapper">
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
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#struct-setup">Struct Setup</a></li>
        <li><a href="#read-function">Read Function</a></li>
        <li><a href="#creating-a-schema">Creating a Schema</a></li>
        <li><a href="#schema-functions">Schema Functions</a></li>
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

Other projects utilizing an SQL wrapper will avoid the headache of setting up new SQL functions for new objects, and/or will be able to quickly change setup in a project instead of rewriting code constantly.

<p align="right">(<a href="#top">back to top</a>)</p>


### Built With

* [Golang](https://go.dev/)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

This wrapper is relatively easy to set up for your projects. These steps will follow along the example project found in the repository.

#### Struct Setup

First off, for each struct you want to wrap, define the struct and the SQL table it represents. At the moment this is a relatively manual process done with struct tags.

```go
type Record struct {
  Author string   `sql:"Author" def:"VARCHAR(128)"`
	Likes  int      `sql:"Likes" def:"INT"`
	Type   PostType `sql:"Type" def:"ENUM('Original', 'Comment', 'Repost')"`
}
```

In this instance, a `Record` struct is created with three different fields (the `Type` field has an enum value defined outside of this snippet). Then, each field has related tags to fit with the definition. These tags are:
* `sql`: the name of the field that will be encoded into the SQL database
* `def`: the SQL definition of the column. This will go directly into a CREATE TABLE statement

### Read Function

Next, a `Read` function needs to be created that is attached to the struct. Because SQL querires need specificity when reading in new values, this is done easiest through a user-defined function. This may change in the future with golang shennanigans, but in the current version you need to define a function specifically to read in new structs.

```go
// Function definition
func (r Record) Read(rows *sql.Rows) (map[int]sql_wrapper.Readable, error) {
  // Create a new list of items
  items := map[int]sql_wrapper.Readable{}

	id := 0 // Note: ID is not a custom variable and is required

  // Custom variables specific to your struct
  author := ""
  likes := 0
  t := Undefined

  // For each row in the pre-existing SQL table
  for rows.Next() {
    if err := rows.Scan(&id, &author, &likes, &t); err != nil {
      return items, err
    }

    // Create a new object and add it to the items
    obj := Record{Author: author, Likes: likes, Type: t}
    items[id] = obj
  }

  return items, nil
}
```

This function is relatively repeatable from project to project. All that differs is how each row is scanned in and what struct is created.

#### Creating a Schema

Now that your struct has been finished, a schema must be created to keep your struct objects synced to SQL. Creating a schema can be done as so with the required parameters.

```go
schema, err := sql_wrapper.NewSchema[Record](db, Record{})
```

In order to create a schema, you must provide:
* The SQL database object needed to execute statements
* An empty struct you are parametrizing the schema with

You can read in existing SQL entries using the `Read` function:

```go
// Read with error handling
if err := schema.Read(); err != nil {
  // handle err
}
```

#### Schema Functions

After your schema is created, you can then call functions associated with it. These are present in the example project and the [documentation][documentation-url].

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [ ] Foriegn Key Constraints
- [ ] Find by ID Functions

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

Project Link: [https://github.com/ethanbaker/sql_wrapper][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/sql_wrapper.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/sql_wrapper.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/sql_wrapper.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/sql_wrapper.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/sql_wrapper.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: <https://github.com/ethanbaker/sql_wrapper/graphs/contributors>
[forks-url]: <https://github.com/ethanbaker/sql_wrapper/network/members>
[stars-url]: <https://github.com/ethanbaker/sql_wrapper/stargazers>
[issues-url]: <https://github.com/ethanbaker/sql_wrapper/issues>
[pulls-url]: <https://github.com/ethanbaker/sql_wrapper/pulls>
[license-url]: <https://github.com/ethanbaker/sql_wrapper/blob/master/LICENSE>
[linkedin-url]: <https://linkedin.com/in/ethandbaker>
[project-url]: <https://github.com/ethanbaker/sql_wrapper>

[documentation-url]: <https://godoc.org/github.com/ethanbaker/sql_wrapper>

[conventional-commits-url]: <https://www.conventionalcommits.org/en/v1.0.0/#summary>
[conventional-branches-url]: <https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops>