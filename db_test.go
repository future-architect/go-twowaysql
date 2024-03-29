package twowaysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestDBConnection(t *testing.T) {
	//データベースは/postgres/init以下のsqlファイルを用いて初期化されている。
	var db *sql.DB
	var err error

	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		db, err = sql.Open("pgx", fmt.Sprintf("host=%s user=postgres password=postgres dbname=postgres sslmode=disable", host))
	} else {
		db, err = sql.Open("pgx", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable")
	}

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT first_name from persons")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		rows.Close()
	})
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Error(err)
		}
		t.Logf("first_name: %v\n", name)
	}
	if err := rows.Err(); err != nil {
		t.Error(err)
	}
}
