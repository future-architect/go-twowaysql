# twowaysql

2-Way-SQL Go implementation

## Installation

```
go get gitlab.com/osaki-lab/twowaysql
```

## Usage

TODO Below is an example which shows some common use cases for twowaysql. 

```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/lib/pq"
    "gitlab.com/osaki-lab/twowaysql"
)

type Person struct {
    FirstName string `db:"first_name"`
    LastName  string `db:"last_name"`
    Email     string `db:"email"`
}

type Params {
    Name       string      `twowaysql:"name"`
    EmpNo      int         `twowaysql:"EmpNo"`
    MaxEmpNo   int         `twowaysql:"maxEmpNo"`
    DeptNo     int         `twowaysql:"deptNo"`
}

func main() {
    ctx := context.Background()

    db, err := twowaysql.Connect("postgres", "user=foo dbname=bar sslmode=disable") 

    var people []Person
    var params = Params{
        MaxEmpNo: 2000,
        deptNp: 15
    }

    err := db.Select(ctx, &people, `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000 /* IF deptNo */ AND dept_no = /*deptNo*/1`, &params)
    if err != nil {
    	log.Fatalf("select failed: %v", err)
    }
    
    fmt.Printf("%#v\n%#v", people[0], people[1])
    // Person{FirstName:"Jason", LastName:"Moiron", Email:"jmoiron@jmoiron.net"}
    // Person{FirstName:"John", LastName:"Doe", Email:"johndoeDNE@gmail.net"}

}
```


## License

Apache License Version 2.0
