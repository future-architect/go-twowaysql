package twowaysql

import (
	"testing"
)

func TestGenerate(t *testing.T) {

	var params = map[string]interface{}{"name": "Jeff", "maxEmpNo": 3, "deptNo": 12}
	tests := []struct {
		name       string
		input      string
		wantQuery  string
		wantParams []interface{}
	}{
		{
			name:      "",
			input:     `SELECT * FROM person WHERE name = /* name */"Tim"`,
			wantQuery: `SELECT * FROM person WHERE name = ?/* name */`,
			wantParams: []interface{}{
				"Jeff",
			},
		},
		{
			name:      "",
			input:     `SELECT * FROM person WHERE empNo < /* maxEmpNo*/100 AND deptNo < /* deptNo */10`,
			wantQuery: `SELECT * FROM person WHERE empNo < ?/* maxEmpNo*/ AND deptNo < ?/* deptNo */`,
			wantParams: []interface{}{
				3,
				12,
			},
		},
		{
			name:      "insert",
			input:     `INSERT INTO persons (employee_no, dept_no, first_name, last_name, email) VALUES(/*maxEmpNo*/1, /*deptNo*/1)`,
			wantQuery: `INSERT INTO persons (employee_no, dept_no, first_name, last_name, email) VALUES(?/*maxEmpNo*/, ?/*deptNo*/)`,
			wantParams: []interface{}{
				3,
				12,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotQuery, gotParams, err := Generate(tt.input, params); err != nil || gotQuery != tt.wantQuery || !interfaceSliceEqual(gotParams, tt.wantParams) {
				if err != nil {
					t.Error(err)
				}
				if gotQuery != tt.wantQuery {
					t.Errorf("\nDoesn't Match\nexpected: \n%s\n but got: \n%s\n", tt.wantQuery, gotQuery)
				}
				if !interfaceSliceEqual(gotParams, tt.wantParams) {
					t.Errorf("\nDoesn't Match\nexpected: \n%v\n but got: \n%v\n", tt.wantParams, gotParams)
				}
			}
		})
	}

}

func interfaceSliceEqual(got, want []interface{}) bool {
	if len(want) != len(got) {
		return false
	}
	for i := 0; i < len(want); i++ {
		if want[i] != got[i] {
			return false
		}
	}
	return true
}
