package twowaysql

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
)

type Person struct {
	FirstName string `twowaysql: "first_name"`
	LastName  string `twowaysql: "last_name"`
	Email     string `twowaysql: "email"`
}

func TestE2E(t *testing.T) {
	//データベースは/postgres/init以下のsqlファイルを用いて初期化されている。
	db, err := sql.Open("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
	defer db.Close()
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	var people []Person
	var params = map[string]interface{}{"maxEmpNo": 3, "deptNo": 12}

	// 式言語に対応していないためif trueとしている
	err = Select(&people, `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000 /* IF false */ AND dept_no = /*deptNo*/'1'`, params).Run(db, ctx)
	if err != nil {
		t.Errorf("select: failed: %v", err)
	}

	expected := []Person{
		{
			FirstName: "Evan",
			LastName:  "MacMans",
			Email:     "evanmacmans@example.com",
		},
		{
			FirstName: "Malvina",
			LastName:  "FitzSimon",
			Email:     "malvinafitzsimons@example.com",
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
