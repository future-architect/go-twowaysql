package twowaysql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func TestDBConnection(t *testing.T) {
	//データベースは/postgres/init以下のsqlファイルを用いて初期化されている。
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
	defer db.Close()
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT first_name from persons")
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Error(err)
		}
		fmt.Printf("first_name: %v\n", name)
	}
	if err := rows.Err(); err != nil {
		t.Error(err)
	}
}
