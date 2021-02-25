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
	db, err := sql.Open("postgres", "user=foo dbname=bar sslmode=disable")
	defer db.Close()
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()

	cmd := `CREATE TABLE IF NOT EXISTS persons(
		employee_no INT,
		dept_no INT,
		first_name STRING,
		last_name STRING,
		email STRING,
		PRIMARY KEY (employee_no)
		)`

	_, err = db.ExecContext(ctx, cmd)
	if err != nil {
		t.Error(err)
	}

	cmd = `INSERT INTO persons (employee_no, dept_no, first_name, last_name, email) VALUES
			(1, 10, 'Evan', 'MacMans', 'evanmacmans@example.com'),
			(2, 11, 'Malvina', 'FitzSimons', 'malvinafitzsimons@examp.com')
			(3, 12, 'Jimmie', 'Bruce', 'jimmiebruce@examp.com')
			`
	_, err = db.ExecContext(ctx, cmd)
	if err != nil {
		t.Error(err)
	}

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
