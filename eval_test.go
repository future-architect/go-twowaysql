package twowaysql

import (
	"database/sql"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

type Info struct {
	Name       string          `twowaysql:"name"`
	EmpNo      int             `twowaysql:"EmpNo"`
	MaxEmpNo   int             `twowaysql:"maxEmpNo"`
	DeptNo     int             `twowaysql:"deptNo"`
	FirstName  string          `twowaysql:"firstName"`
	LastName   string          `twowaysql:"lastName"`
	Email      string          `twowaysql:"email"`
	GenderList []string        `twowaysql:"gender_list"`
	IntList    []int           `twowaysql:"int_list"`
	Checked    bool            `twowaysql:"checked"`
	Unchecked  bool            `twowaysql:"unchecked"`
	Nil        interface{}     `twowaysql:"nil"`
	Zero       int             `twowaysql:"zero"`
	Table      [][]interface{} `twowaysql:"table"`
	NullString sql.NullString  `twowaysql:"null_string"`
	NullInt    sql.NullInt64   `twowaysql:"null_int"`
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
			name:  "in bind table",
			input: `SELECT * FROM person WHERE employee_no = /*maxEmpNo*/1000 /* IF table !== undefined */ AND (person.gender, person.Name) in /*table*/(('M', 'Jeff'), ('F', 'Jeff')) /* END */`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
				Table: [][]interface{}{
					{"M", "Tom"},
					{"F", "Tom"},
				},
			},
			wantQuery: `SELECT * FROM person WHERE employee_no = ?/*maxEmpNo*/ AND (person.gender, person.Name) in ((?, ?), (?, ?))/*table*/`,
			wantParams: []interface{}{
				3,
				"M",
				"Tom",
				"F",
				"Tom",
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
		{
			name: "multiline if true",
			input: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no < 1000
				/* IF true */
				AND dept_no = 1
				/* END */
			`,
			inputParams: Info{},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no < 1000
				
				AND dept_no = 1
				
			`,
			wantParams: []interface{}{},
		},
		{
			name: "multiline if false elif false else",
			input: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no < 1000
				/*
					IF false
				*/
				AND dept_no = 1
				/*
					ELIF false
				*/
				AND boss_no = 2
				/*
					ELSE
				*/
				AND id = /*maxEmpNo*/3
				/*
					END
				*/
			`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no < 1000
				
				AND id = ?/*maxEmpNo*/
				
			`,
			wantParams: []interface{}{3},
		},
		{
			name: "multiline if nest",
			input: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no	<	1000
				/*
					IF true
				*/
					/*
						IF false
					*/
					AND	dept_no		=	1
					/*
						ELSE
					*/
					AND	id			=	3
					/*
						END
					*/
				/*
					ELSE
				*/
				AND	boss_id		=	4
				/*
					END
				*/
			`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no	<	1000
				
					
					AND	id			=	3
					
				
			`,
			wantParams: []interface{}{},
		},
		{
			name: "multiline in bind string",
			input: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no		=	/*maxEmpNo*/1000
				/*
					IF int_list !== null
				*/
				AND	person.gender	in	/*int_list*/(3, 5, 7)
				/*
					END
				*/
			`,
			inputParams: Info{
				Name:       "Jeff",
				MaxEmpNo:   3,
				DeptNo:     12,
				GenderList: []string{"M", "F"},
				IntList:    []int{1, 2, 3},
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE
				employee_no		=	?/*maxEmpNo*/
				
				AND	person.gender	in	(?, ?, ?)/*int_list*/
				
			`,
			wantParams: []interface{}{
				3,
				1,
				2,
				3,
			},
		},
		{
			name: "multiple if condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF EmpNo !== null */
				 AND employee_no <   /*EmpNo*/'0001'
				/* END */
				/* IF maxEmpNo !== null */
				 AND id =   /*maxEmpNo*/'0002'
				/* END */
				/* IF deptNo !== null */
				 AND dept_no =   /*deptNo*/1
				/* END */
				/* IF firstName !== null */
				 AND first_name =   /*firstName*/'first_name'
				/* END */
				/* IF lastName !== null */
				 AND last_name =   /*lastName*/'last_name'
				/* END */
				/* IF email !== null */
				 AND email =   /*email*/'email'
				/* END */
				/* IF gender_list !== null */
				 AND gender_list IN /*gender_list*/('01', '02', '03')
				/* END */
			`,
			inputParams: Info{
				EmpNo:      1000,
				MaxEmpNo:   10,
				DeptNo:     1,
				FirstName:  "first",
				LastName:   "last",
				Email:      "email",
				GenderList: []string{"f", "m", "o"},
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
				 AND employee_no < ?/*EmpNo*/
				
				
				 AND id = ?/*maxEmpNo*/
				
				
				 AND dept_no = ?/*deptNo*/
				
				
				 AND first_name = ?/*firstName*/
				
				
				 AND last_name = ?/*lastName*/
				
				
				 AND email = ?/*email*/
				
				
				 AND gender_list IN (?, ?, ?)/*gender_list*/
				
			`,
			wantParams: []interface{}{1000, 10, 1, "first", "last", "email", "f", "m", "o"},
		},
		{
			name: "multiple if false condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF false */
				 AND employee_no <   /*EmpNo*/'0001'
				/* END */
				/* IF maxEmpNo !== null */
				 AND id =   /*maxEmpNo*/'0002'
				/* END */
			`,
			inputParams: Info{
				EmpNo:    1000,
				MaxEmpNo: 10,
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
				
				 AND id = ?/*maxEmpNo*/
				
			`,
			wantParams: []interface{}{10},
		},
		{
			name: "multiple nest if condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF name !== null */
					/* IF EmpNo !== null */
						AND employee_no <   /*EmpNo*/'0001'
					/* ELSE */
						AND employee_no < 1000
					/* END */
					/* IF maxEmpNo !== null */
						AND id =   /*maxEmpNo*/'0002'
					/* ELSE */
						AND employee_no < 1000
					/* END */
				/* ELIF false */
					AND employee_no < 100
				/* ELSE */
					AND employee_no < 1000
				/* END */
			`,
			inputParams: Info{
				Name:     "x",
				EmpNo:    1000,
				MaxEmpNo: 10,
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
					
						AND employee_no < ?/*EmpNo*/
					
					
						AND id = ?/*maxEmpNo*/
					
				
			`,
			wantParams: []interface{}{1000, 10},
		},
		{
			name: "multiple nest elif condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF false */
					AND employee_no < 100
				/* ELIF true */
					/* IF EmpNo !== null */
						AND employee_no <   /*EmpNo*/'0001'
					/* ELSE */
						AND employee_no < 1000
					/* END */
					/* IF maxEmpNo !== null */
						AND id =   /*maxEmpNo*/'0002'
					/* ELSE */
						AND employee_no < 1000
					/* END */
				/* ELSE */
					AND employee_no < 1000
				/* END */
			`,
			inputParams: Info{
				Name:     "x",
				EmpNo:    1000,
				MaxEmpNo: 10,
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
					
						AND employee_no < ?/*EmpNo*/
					
					
						AND id = ?/*maxEmpNo*/
					
				
			`,
			wantParams: []interface{}{1000, 10},
		},
		{
			name: "multiple nest if condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF false */
					AND employee_no < 100
				/* ELIF false */
					AND employee_no < 1000
				/* ELSE */
					/* IF EmpNo !== null */
						AND employee_no <   /*EmpNo*/'0001'
					/* ELSE */
						AND employee_no < 1000
					/* END */
					/* IF maxEmpNo !== null */
						AND id =   /*maxEmpNo*/'0002'
					/* ELSE */
						AND employee_no < 1000
					/* END */
				/* END */
			`,
			inputParams: Info{
				Name:     "x",
				EmpNo:    1000,
				MaxEmpNo: 10,
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
					
						AND employee_no < ?/*EmpNo*/
					
					
						AND id = ?/*maxEmpNo*/
					
				
			`,
			wantParams: []interface{}{1000, 10},
		},
		{
			name: "multiple 4 nest if condition",
			input: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				/* IF name !== null */
					/* IF name !== null */
						/* IF name !== null */
							/* IF false */
								2=2
							/* ELSE */
								/* IF EmpNo === null */
									AND employee_no <   /*EmpNo*/'0001'
								/* ELSE */
									AND employee_no < 222
								/* END */
								/* IF false */
									AND id =   1
								/* ELIF maxEmpNo !== null */
									AND id =   /*maxEmpNo*/'0002'
								/* END */
							/* END */
						/* END */
					/* END */
				/* END */
			`,
			inputParams: Info{
				Name:     "x",
				EmpNo:    1000,
				MaxEmpNo: 10,
			},
			wantQuery: `
			SELECT
				*
			FROM
				person
			WHERE	1=1
				
					
						
							
								
									AND employee_no < 222
								
								
									AND id = ?/*maxEmpNo*/
								
							
						
					
				
			`,
			wantParams: []interface{}{10},
		},
		{
			name: "comment",
			input: `
			-- header comment
			SELECT -- inner comment
				* -- inner comment
			FROM -- inner comment
				person -- inner comment
			WHERE -- inner comment
				employee_no < 1000 -- inner comment
				/* IF true */ -- inner comment
				AND dept_no = 1 -- inner comment
				/* END */ -- inner comment
			-- footer comment
			`,
			inputParams: Info{},
			wantQuery: `
			-- header comment
			SELECT -- inner comment
				* -- inner comment
			FROM -- inner comment
				person -- inner comment
			WHERE -- inner comment
				employee_no < 1000 -- inner comment
				 -- inner comment
				AND dept_no = 1 -- inner comment
				 -- inner comment
			-- footer comment
			`,
			wantParams: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params, err := Eval(tt.input, &tt.inputParams)
			assert.NilError(t, err)
			assert.Check(t, cmp.DeepEqual(tt.wantParams, params))
			assert.Check(t, cmp.DeepEqual(tt.wantQuery, query))
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
			wantError: "can not parse: not found /* END */",
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

func TestEvalWithMap(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		inputParams map[string]interface{}
		wantQuery   string
		wantParams  []interface{}
	}{
		{
			name:  "bind parameter",
			input: `SELECT * FROM person WHERE employee_no < /*maxEmpNo*/1000`,
			inputParams: map[string]interface{}{
				"name":       "Jeff",
				"maxEmpNo":   3,
				"deptNo":     12,
				"genderList": []string{"M", "F"},
				"intList":    []int{1, 2, 3},
			},
			wantQuery:  `SELECT * FROM person WHERE employee_no < ?/*maxEmpNo*/`,
			wantParams: []interface{}{3},
		},
		{
			name:  "bind nil parameter",
			input: `SELECT * FROM person WHERE 1=1 /* IF genderList !== undefined */ AND gender_list IN /*genderList*/('M', 'F') /* END */`,
			inputParams: map[string]interface{}{
				"genderList": nil,
			},
			wantQuery:  `SELECT * FROM person WHERE 1=1`,
			wantParams: []interface{}{},
		},
		{
			name:  "bind table parameter",
			input: `SELECT * FROM person WHERE name = /*name*/'name' AND (a, b) IN /*table*/(('x', 10), ('y', 11))`,
			inputParams: map[string]interface{}{
				"name": "Jeff",
				"table": [][]interface{}{
					{"a", 1},
					{"b", 2},
					{"c", 3},
				},
			},
			wantQuery:  `SELECT * FROM person WHERE name = ?/*name*/ AND (a, b) IN ((?, ?), (?, ?), (?, ?))/*table*/`,
			wantParams: []interface{}{"Jeff", "a", 1, "b", 2, "c", 3},
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

func TestEval_SQLNullTyp(t *testing.T) {
	type SQLTypInfo struct {
		NullBool    sql.NullBool    `db:"null_bool"`
		NullFloat64 sql.NullFloat64 `db:"null_float_64"`
		NullInt16   sql.NullInt16   `db:"null_int_16"`
		NullInt32   sql.NullInt32   `db:"null_int_32"`
		NullInt64   sql.NullInt64   `db:"null_int_64"`
		NullString  sql.NullString  `db:"null_string"`
		NullTime    sql.NullTime    `db:"null_time"`
	}

	tests := []struct {
		name        string
		input       string
		inputParams SQLTypInfo
		wantQuery   string
		wantParams  []interface{}
	}{
		{
			name:  "bind sql.NullBool",
			input: `SELECT * FROM person WHERE value = /*null_bool*/false`,
			inputParams: SQLTypInfo{
				NullBool: sql.NullBool{Bool: true, Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_bool*/`,
			wantParams: []interface{}{true},
		},
		{
			name:  "bind sql.Float64",
			input: `SELECT * FROM person WHERE value = /*null_float_64*/1.0`,
			inputParams: SQLTypInfo{
				NullFloat64: sql.NullFloat64{Float64: 10.01, Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_float_64*/`,
			wantParams: []interface{}{10.01},
		},
		{
			name:  "bind sql.NullInt16",
			input: `SELECT * FROM person WHERE value = /*null_int_16*/1`,
			inputParams: SQLTypInfo{
				NullInt16: sql.NullInt16{Int16: 10, Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_int_16*/`,
			wantParams: []interface{}{int16(10)},
		},
		{
			name:  "bind sql.NullInt32",
			input: `SELECT * FROM person WHERE value = /*null_int_32*/1`,
			inputParams: SQLTypInfo{
				NullInt32: sql.NullInt32{Int32: 100, Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_int_32*/`,
			wantParams: []interface{}{int32(100)},
		},
		{
			name:  "bind sql.NullInt64",
			input: `SELECT * FROM person WHERE value = /*null_int_64*/1`,
			inputParams: SQLTypInfo{
				NullInt64: sql.NullInt64{Int64: 1000, Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_int_64*/`,
			wantParams: []interface{}{int64(1000)},
		},
		{
			name:  "bind sql.NullString",
			input: `SELECT * FROM person WHERE value = /*null_string*/'hoge'`,
			inputParams: SQLTypInfo{
				NullString: sql.NullString{String: "value", Valid: true},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_string*/`,
			wantParams: []interface{}{"value"},
		},
		{
			name:  "bind sql.NullTime",
			input: `SELECT * FROM person WHERE value = /*null_time*/'2022-01-01 10:00:00'`,
			inputParams: SQLTypInfo{
				NullTime: sql.NullTime{
					Time:  time.Date(2022, 7, 1, 12, 30, 30, 0, time.UTC),
					Valid: true,
				},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_time*/`,
			wantParams: []interface{}{time.Date(2022, 7, 1, 12, 30, 30, 0, time.UTC)},
		},
		{
			name:        "bind initial",
			input:       `SELECT * FROM person WHERE value = /*null_string*/'hoge'`,
			inputParams: SQLTypInfo{},
			wantQuery:   `SELECT * FROM person WHERE value = ?/*null_string*/`,
			wantParams:  []interface{}{nil},
		},
		{
			name:  "bind invalid",
			input: `SELECT * FROM person WHERE value = /*null_string*/'hoge'`,
			inputParams: SQLTypInfo{
				NullString: sql.NullString{String: "value", Valid: false},
			},
			wantQuery:  `SELECT * FROM person WHERE value = ?/*null_string*/`,
			wantParams: []interface{}{nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, params, err := Eval(tt.input, &tt.inputParams)
			assert.NilError(t, err)
			assert.Check(t, cmp.DeepEqual(tt.wantParams, params))
			assert.Check(t, cmp.DeepEqual(tt.wantQuery, query))
		})
	}
}
