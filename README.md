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
    FirstName string `twowaysql:"first_name"`
    LastName  string `twowaysql:"last_name"`
    Email     string
}

func main() {
    ctx := context.Background()

    db, err := twowaysql.Connect("postgres", "user=foo dbname=bar sslmode=disable") 

    var people []Person
    var params = map[string]interface{}{"maxEmpNo": 2000, "deptNo":15} 
    err := db.Select(&people, `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000 /* IF exists(deptNo)*/ AND dept_no = /*deptNo*/'1'`).Run(ctx, params)
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
