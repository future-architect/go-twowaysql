package twowaysql_test

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/osaki-lab/twowaysql"
)

var (
	tw  *twowaysql.Twowaysql
	ctx context.Context
)

func ExampleTwowaysql_Exec() {

	type Info struct {
		Name       string   `map:"name"`
		EmpNo      int      `map:"EmpNo"`
		MaxEmpNo   int      `map:"maxEmpNo"`
		DeptNo     int      `map:"deptNo"`
		Email      string   `map:"email"`
		GenderList []string `map:"gender_list"`
		IntList    []int    `map:"int_list"`
	}

	var params = Info{
		MaxEmpNo: 3,
		DeptNo:   12,
	}
	result, err := tw.Exec(ctx, `UPDATE persons SET dept_no = /*deptNo*/1 WHERE employee_no = /*EmpNo*/1`, &params)
	if err != nil {
		log.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows != 1 {
		log.Fatalf("expected to affect 1 row. affected %d", rows)
	}
}

func ExampleTwowaysql_Select() {
	type Person struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}

	type Info struct {
		Name       string   `map:"name"`
		EmpNo      int      `map:"EmpNo"`
		MaxEmpNo   int      `map:"maxEmpNo"`
		DeptNo     int      `map:"deptNo"`
		Email      string   `map:"email"`
		GenderList []string `map:"gender_list"`
		IntList    []int    `map:"int_list"`
	}

	var people []Person
	var params = Info{
		MaxEmpNo: 3,
		DeptNo:   12,
	}

	err := tw.Select(ctx, &people, `SELECT first_name, last_name, email FROM persons WHERE employee_no < /*maxEmpNo*/1000 /* IF deptNo */ AND dept_no < /*deptNo*/1 /* END */`, &params)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleEval() {

	type Info struct {
		Name       string   `map:"name"`
		EmpNo      int      `map:"EmpNo"`
		MaxEmpNo   int      `map:"maxEmpNo"`
		DeptNo     int      `map:"deptNo"`
		Email      string   `map:"email"`
		GenderList []string `map:"gender_list"`
		IntList    []int    `map:"int_list"`
	}

	var params = Info{
		Name:       "Jeff",
		MaxEmpNo:   3,
		DeptNo:     12,
		GenderList: []string{"M", "F"},
		IntList:    []int{1, 2, 3},
	}
	var query = `SELECT * FROM person WHERE employee_no = /*maxEmpNo*/1000 AND /* IF int_list !== null */  person.gender in /*int_list*/(3,5,7) /* END */`

	convertedQuery, getParams, _ := twowaysql.Eval(query, &params)

	fmt.Println(convertedQuery)
	fmt.Println(getParams)

	// Output:
	// SELECT * FROM person WHERE employee_no = ?/*maxEmpNo*/ AND person.gender in (?, ?, ?)/*int_list*/
	// [3 1 2 3]

}
