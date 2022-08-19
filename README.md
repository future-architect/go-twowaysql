# twowaysql
![test](https://github.com/future-architect/go-twowaysql/actions/workflows/test.yml/badge.svg)


2-Way-SQL Go implementation

## Installation

```
go get github.com/future-architect/go-twowaysql 
```

## Usage

TODO Below is an example which shows some common use cases for twowaysql. 

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/future-architect/go-twowaysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type Person struct {
	EmpNo     int    `db:"employee_no"`
	DeptNo    int    `db:"dept_no"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}

type Params struct {
	Name     string `twowaysql:"name"`
	EmpNo    int    `twowaysql:"EmpNo"`
	MaxEmpNo int    `twowaysql:"maxEmpNo"`
	DeptNo   int    `twowaysql:"deptNo"`
}

func main() {
	ctx := context.Background()

	db, err := sqlx.Open("pgx", "user=postgres password=postgres dbname=postgres sslmode=disable")

	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	tw := twowaysql.New(db)

	var people []Person
	var params = Params{
		MaxEmpNo: 2000,
		DeptNo:   15,
	}

	err = tw.Select(ctx, &people, `SELECT * FROM persons WHERE employee_no < /*maxEmpNo*/1000 /* IF deptNo */ AND dept_no < /*deptNo*/1 /* END */`, &params)
	if err != nil {
		log.Fatalf("select failed: %v", err)
	}

	fmt.Printf("%#v\n%#v\n%#v", people[0], people[1], people[2])
	//Person{EmpNo:1, DeptNo:10, FirstName:"Evan", LastName:"MacMans", Email:"evanmacmans@example.com"}
	//Person{EmpNo:3, DeptNo:12, FirstName:"Jimmie", LastName:"Bruce", Email:"jimmiebruce@example.com"}
	//Person{EmpNo:2, DeptNo:11, FirstName:"Malvina", LastName:"FitzSimons", Email:"malvinafitzsimons@example.com"}

}

```

## License

Apache License Version 2.0

## Contribution

Launch database for testing:

```
$ docker compose up --build
```

Run acceptance test:

```
$ docker compose -f docker-compose-test.yml up --build
```