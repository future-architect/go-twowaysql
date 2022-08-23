package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "modernc.org/sqlite"

	// _ "github.com/lib/pq"

	"github.com/future-architect/go-twowaysql/cli"
)

func main() {
	cli.Main()
}
