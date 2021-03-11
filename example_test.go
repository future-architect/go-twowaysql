package twowaysql_test

import (
	"fmt"

	"gitlab.com/osaki-lab/twowaysql"
)

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
