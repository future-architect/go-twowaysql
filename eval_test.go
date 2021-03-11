package twowaysql

import (
	"testing"
)

type Info struct {
	Name       string      `twowaysql:"name"`
	EmpNo      int         `twowaysql:"EmpNo"`
	MaxEmpNo   int         `twowaysql:"maxEmpNo"`
	DeptNo     int         `twowaysql:"deptNo"`
	FirstName  string      `twowaysql:"firstName"`
	LastName   string      `twowaysql:"lastName"`
	Email      string      `twowaysql:"email"`
	GenderList []string    `twowaysql:"gender_list"`
	IntList    []int       `twowaysql:"int_list"`
	Checked    bool        `twowaysql:"checked"`
	Unchecked  bool        `twowaysql:"unchecked"`
	Nil        interface{} `twowaysql:"nil"`
	Zero       int         `twowaysql:"zero"`
}

func TestEval(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		inputParams Info
		wantQuery   string
		wantParams  []interface{}
	}{
		{
			name:        "if true",
			input:       `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if false",
			input:       `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if true else",
			input:       `SELECT * FROM person WHERE employee_no < 1000  /* IF true */ AND dept_no = 1 /* ELSE */ boss_no = 2 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if false else",
			input:       `SELECT * FROM person WHERE employee_no < 1000  /* IF false */ AND  dept_no =1 /* ELSE */AND boss_no = 2 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND boss_no = 2`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if true elif true else",
			input:       `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1 /* ELIF true */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if false elif true else",
			input:       `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF true */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND boss_no = 2`,
			wantParams:  []interface{}{},
		},
		{
			name:        "if false elif false else",
			input:       `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF false */ AND boss_no = 2 /* ELSE */ AND id = 3 /* END */`,
			inputParams: Info{},
			wantQuery:   `SELECT * FROM person WHERE employee_no < 1000 AND id = 3`,
			wantParams:  []interface{}{},
		},
		{
			name:  "bind parameter",
			input: `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery:  `SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/`,
			wantParams: []interface{}{3},
		},
		{
			name:  "if false elif false else",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF false */ AND dept_no = 1 /* ELIF false */ AND boss_no = 2 /* ELSE */ AND id = /*maxEmpNo*/3 /* END */`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery:  `SELECT * FROM person WHERE employee_no < 1000 AND id = ?/*maxEmpNo*/`,
			wantParams: []interface{}{3},
		},
		{
			name:  "if nest",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ /* IF false */ AND dept_no =1 /* ELSE */ AND id=3 /* END */ /* ELSE*/ AND boss_id=4 /* END */`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery:  `SELECT * FROM person WHERE employee_no < 1000 AND id=3`,
			wantParams: []interface{}{},
		},
		{
			name:  "bind string",
			input: `SELECT * FROM person WHERE name = /* name */"Tim"`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `SELECT * FROM person WHERE name = ?/* name */`,
			wantParams: []interface{}{
				"Jeff",
			},
		},
		{
			name:  "bind ints",
			input: `SELECT * FROM person WHERE empNo < /* maxEmpNo*/100 AND deptNo < /* deptNo */10`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `SELECT * FROM person WHERE empNo < ?/* maxEmpNo*/ AND deptNo < ?/* deptNo */`,
			wantParams: []interface{}{
				3,
				12,
			},
		},
		{
			name:  "insert",
			input: `INSERT INTO persons (employee_no, dept_no, first_name, last_name, email) VALUES(/*maxEmpNo*/1, /*deptNo*/1)`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `INSERT INTO persons (employee_no, dept_no, first_name, last_name, email) VALUES(?/*maxEmpNo*/, ?/*deptNo*/)`,
			wantParams: []interface{}{
				3,
				12,
			},
		},
		{
			name:  "in bind string",
			input: `SELECT * FROM person /* IF gender_list !== null */ WHERE person.gender in /*gender_list*/('M') /* END */`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `SELECT * FROM person WHERE person.gender in (?, ?)/*gender_list*/`,
			wantParams: []interface{}{
				"M",
				"F",
			},
		},
		{
			name:  "in bind string",
			input: `SELECT * FROM person WHERE employee_no = /*maxEmpNo*/1000 /* IF int_list !== null */ AND  person.gender in /*int_list*/(3,5,7) /* END */`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `SELECT * FROM person WHERE employee_no = ?/*maxEmpNo*/ AND person.gender in (?, ?, ?)/*int_list*/`,
			wantParams: []interface{}{
				3,
				1,
				2,
				3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if query, params, err := Eval(tt.input, &tt.inputParams); err != nil || query != tt.wantQuery || !interfaceSliceEqual(params, tt.wantParams) {
				if err != nil {
					t.Error(err)
				}
				if query != tt.wantQuery {
					t.Errorf("Doesn't Match\nexpected: \n%s\n but got: \n%s\n", tt.wantQuery, query)
				}
				if !interfaceSliceEqual(params, tt.wantParams) {
					t.Errorf("Doesn't Match\nexpected: \n%v\n but got: \n%v\n", tt.wantParams, params)
				}
			}
		})
	}
}

func TestCondition(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		inputParams Info
		want        string
	}{
		{
			name:  "if true",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF checked */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if false",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF unchecked */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if truthy 1",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if truthy 2",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if truthy 3",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if falsy 1",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF zero */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if falsy 2",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF nil */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if equal true int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo === 15 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if equal false int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo === 10 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if not equal true int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF maxEmpNo !== 15 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if not equal false int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF maxEmpNo !== 2000 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if equal true string",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name === "HR" */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if equal false string",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF name === "GA" */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if less than true int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo < 100 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if less than false int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo < 10 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
		{
			name:  "if more than true int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo > 10 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000 AND dept_no = 1`,
		},
		{
			name:  "if more than false int",
			input: `SELECT * FROM person WHERE employee_no < 1000 /* IF deptNo > 100 */ AND dept_no = 1 /* END */`,
			inputParams: Info{
				Name:      "HR",
				MaxEmpNo:  2000,
				DeptNo:    15,
				Checked:   true,
				Unchecked: false,
				Zero:      0,
				Nil:       nil,
			},
			want: `SELECT * FROM person WHERE employee_no < 1000`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if query, _, err := Eval(tt.input, &tt.inputParams); err != nil || query != tt.want {
				if err != nil {
					t.Error(err)
				}
				t.Errorf("Doesn't Match\nexpected: \n%s\n but got: \n%s\n", tt.want, query)
			}
		})
	}
}

func TestGenerateAbnormal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "no END",
			input:     `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ AND dept_no = 1`,
			wantError: "can not parse: expected /* END */, but got 7",
		},
		{
			name:      "extra END 1",
			input:     "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1 /* END */",
			wantError: "can not generate abstract syntax tree",
		},
		{
			name:      "extra END 2",
			input:     "SELECT * FROM person WHERE employee_no < 1000  /* END */ AND dept_no = 1 ",
			wantError: "can not generate abstract syntax tree",
		},
		{
			name:      "invalid Elif pos",
			input:     `SELECT * FROM person WHERE employee_no < 1000 /* ELIF true */ AND dept_no = 1`,
			wantError: "can not generate abstract syntax tree",
		},
		{
			name:      "not match if, elif and end",
			input:     `SELECT * FROM person WHERE employee_no < 1000 /* IF true */ /* IF false */ AND dept_no =1 /* ELSE */ AND id=3 /* ELSE*/ AND boss_id=4 /* END */`,
			wantError: "can not parse: expected /* END */, but got 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if query, params, err := Eval(tt.input, nil); err == nil || err.Error() != tt.wantError {
				if err == nil {
					t.Error("query", query)
					t.Error("params", params)
					t.Errorf("should return error")
				} else {
					t.Errorf("\nexpected:\n%v\nbut got\n%v\n", tt.wantError, err.Error())
				}
			}
		})
	}
}

func stringSliceEqual(want, got []interface{}) bool {
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
