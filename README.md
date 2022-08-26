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

## CLI Tool

CLI tool `twowaysql` provides helper functions about two way sql

```
go install github.com/future-architect/go-twowaysql/...
```

### Database Connection

To connect database, *driver* and *source* strings are required. Driver is like `pgx` and source is `postgres://user:pass@host/dbname?sslmode=disable`.

You can pass them via options(`-d DRIVER`, `--driver=DRIVER`, `-c SOURCE`, `--source=SOURCE`) or by using `TWOWAYSQL_DRIVER`/`TWOWAYSQL_CONNECTION` environment variables.

This tool also read `.env` and `.env.local` files.

### Execute SQL

```
$ twowaysql run -p first_name=Malvina testdata/postgres/sql/select_person.sql
┌───────────────────────────────┬────────────┬────────────┐
│ email                         │ first_name │ last_name  │
╞═══════════════════════════════╪════════════╪════════════╡
│ malvinafitzsimons@example.com │ Malvina    │ FitzSimons │
└───────────────────────────────┴────────────┴────────────┘

Query takes 22.804166ms
```

* -p, --param=PARAM ...        Parameter in single value or JSON (name=bob, or {"name": "bob"})
* -e, --explain                Run with EXPLAIN to show execution plan
* -r, --rollback               Run within transaction and then rollback
* -o, --output-format=default  Result output format (default, md, json, yaml)

### Evaluate 2-Way-SQL

```
$ twowaysql eval -p first_name=Malvina testdata/postgres/sql/select_person.sql
# Converted Source

SELECT email, first_name, last_name FROM persons WHERE first_name=?/*first_name*/;

# Parameters

- Malvina
```

### Customize CLI tool

by default `twowaysql` integrated with the following drivers:

* ``github.com/jackc/pgx/v4``
* ``modernc.org/sqlite``
* ``github.com/go-sql-driver/mysql``

If you want to add/remove [drivers](https://github.com/golang/go/wiki/SQLDrivers), create simple main package and call `cli.Main()`.

```go
package main

import (
	_ "github.com/sijms/go-ora/v2" // Oracle

	"github.com/future-architect/go-twowaysql/cli"
)

func main() {
	cli.Main()
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