package twowaysql_test

import (
	"fmt"

	"gitlab.com/osaki-lab/twowaysql"
)

func ExampleGenerate() {

	var params = map[string]interface{}{"name": "Jeff", "maxEmpNo": 3, "deptNo": 12, "gender_list": []string{"M", "F"}, "int_list": []int{1, 2, 3}}
	var query = `SELECT * FROM person WHERE employee_no = /*maxEmpNo*/1000 AND /* IF int_list !== null */  person.gender in /*int_list*/(3,5,7) /* END */`

	convertedQuery, getParams, _ := twowaysql.Generate(query, params)

	fmt.Println("query:", convertedQuery)
	fmt.Println("params:", getParams)

	// Output:
	// query: SELECT * FROM person WHERE employee_no = ?/*maxEmpNo*/ AND person.gender in (?, ?, ?)/*int_list*/
	// params: [3 1 2 3]

}
