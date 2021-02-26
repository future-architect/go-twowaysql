package twowaysql

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

func TestSelect(t *testing.T) {
	//データベースは/postgres/init以下のsqlファイルを用いて初期化されている。
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
	defer db.Close()
	if err != nil {
		t.Error(err)
	}

	tw := New(db)

	ctx := context.Background()

	var people []Person
	var params = map[string]interface{}{"maxEmpNo": 3, "deptNo": 10}

	// 式言語に対応していないためif trueとしている
	err = tw.Select(&people, `SELECT first_name, last_name, email FROM persons WHERE employee_no < /*maxEmpNo*/1000 /* IF deptNo */ AND dept_no = /*deptNo*/'1' /* END */`, params).Run(ctx)
	if err != nil {
		t.Errorf("select: failed: %v", err)
	}

	expected := []Person{
		{
			FirstName: "Evan",
			LastName:  "MacMans",
			Email:     "evanmacmans@example.com",
		},
	}

	if !match(people, expected) {
		t.Errorf("expected:\n%v\nbut got\n%v\n", expected, people)
	}
}

func match(p1, p2 []Person) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := 0; i < len(p1); i++ {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}
