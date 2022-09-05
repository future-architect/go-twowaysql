package twowaysql

import (
	"fmt"
	"testing"
)

func TestAst(t *testing.T) {
	tests := []struct {
		name  string
		input []token
		want  *tree
	}{
		{
			name: "empty",
			input: []token{
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantEmpty,
		},
		{
			name: "no comment",
			input: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantNoComment,
		},
		{
			name: "if",
			input: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind:      tkIf,
					str:       "/* IF true */",
					condition: "true",
				},
				{
					kind: tkSQLStmt,
					str:  " AND dept_no = 1",
				},
				{
					kind: tkEnd,
					str:  "/* END */",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantTreeIf,
		},
		{
			name: "if and bind",
			input: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < ",
				},
				{
					kind:  tkBind,
					str:   "?/*maxEmpNo*/",
					value: "maxEmpNo",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
				{
					kind:      tkIf,
					str:       "/* IF false */",
					condition: "false",
				},
				{
					kind: tkSQLStmt,
					str:  " AND dept_no = ",
				},
				{
					kind:  tkBind,
					str:   "?/*deptNo*/",
					value: "deptNo",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
				{
					kind: tkEnd,
					str:  "/* END */",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantTreeIfBind,
		},
		{
			name: "if elif else",
			input: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind:      tkIf,
					str:       "/* IF true */",
					condition: "true",
				},
				{
					kind: tkSQLStmt,
					str:  "AND dept_no =1",
				},
				{
					kind:      tkElif,
					str:       "/* ELIF true*/",
					condition: "true",
				},
				{
					kind: tkSQLStmt,
					str:  " AND boss_no = 2 ",
				},
				{
					kind: tkElse,
					str:  "/*ELSE */",
				},
				{
					kind: tkSQLStmt,
					str:  " AND id=3",
				},
				{
					kind: tkEnd,
					str:  "/* END */",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantIfElifElse,
		},
		{
			name: "if nest",
			input: []token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind:      tkIf,
					str:       "/* IF true */",
					condition: "true",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
				{
					kind:      tkIf,
					str:       "/* IF false */",
					condition: "false",
				},
				{
					kind: tkSQLStmt,
					str:  " AND dept_no =1 ",
				},
				{
					kind: tkElse,
					str:  "/* ELSE */",
				},
				{
					kind: tkSQLStmt,
					str:  " AND id=3 ",
				},
				{
					kind: tkEnd,
					str:  "/* END */",
				},
				{
					kind: tkSQLStmt,
					str:  " ",
				},
				{
					kind: tkElse,
					str:  "/* ELSE*/",
				},
				{
					kind: tkSQLStmt,
					str:  " AND boss_id=4 ",
				},
				{
					kind: tkEnd,
					str:  "/* END */",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: wantIfNest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ast(tt.input); err != nil || !treeEqual(tt.want, got) {
				if err != nil {
					t.Log(err)
				}
				t.Errorf("Doesn't Match expected: %v, but got: %v\n", tt.want, got)
				fmt.Println("want:")
				printTree(tt.want)
				fmt.Println("got:")
				printTree(got)
			}
		})
	}

}

func walk(t *tree, ch chan *token) {
	defer close(ch)

	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)

}

func walkInner(t *tree, ch chan *token) {
	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)
}

func treeEqual(t1, t2 *tree) bool {
	ch1 := make(chan *token)
	ch2 := make(chan *token)

	go walk(t1, ch1)
	go walk(t2, ch2)

	var s1, s2 []*token

	for n := range ch1 {
		s1 = append(s1, n)
	}
	for n := range ch2 {
		s2 = append(s2, n)
	}

	if len(s1) == len(s2) {
		for i := 0; i < len(s1); i++ {
			if *s1[i] != *s2[i] {
				fmt.Printf("not macth %v != %v\n", s1[i], s2[i])
				return false
			}
		}
		return true
	}
	return false
}

func printTree(t *tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)

}

func printWalkInner(t *tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)
}

// テストの期待する結果を作成
var (
	wantEmpty = &tree{
		Kind: ndEndOfProgram,
		Token: &token{
			kind: tkEndOfProgram,
		},
	}

	wantNoComment = &tree{
		Kind: ndSQLStmt,
		Left: &tree{
			Kind: ndEndOfProgram,
			Token: &token{
				kind: tkEndOfProgram,
			},
		},
		Token: &token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
		},
	}

	wantTreeIf = &tree{
		Kind: ndSQLStmt,
		Left: &tree{
			Kind: ndIf,
			Left: &tree{
				Kind: ndSQLStmt,
				Token: &token{
					kind: tkSQLStmt,
					str:  " AND dept_no = 1",
				},
			},
			Right: &tree{
				Kind: ndEnd,
				Left: &tree{
					Kind: ndEndOfProgram,
					Token: &token{
						kind: tkEndOfProgram,
					},
				},
				Token: &token{
					kind: tkEnd,
					str:  "/* END */",
				},
			},
			Token: &token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	wantTreeIfBind = &tree{
		Kind: ndSQLStmt,
		Left: &tree{
			Kind: ndBind,
			Left: &tree{
				Kind: ndSQLStmt,
				Left: &tree{
					Kind: ndIf,
					Left: &tree{
						Kind: ndSQLStmt,
						Left: &tree{
							Kind: ndBind,
							Left: &tree{
								Kind: ndSQLStmt,
								Token: &token{
									kind: tkSQLStmt,
									str:  " ",
								},
							},
							Token: &token{
								kind:  tkBind,
								str:   "?/*deptNo*/",
								value: "deptNo",
							},
						},
						Token: &token{
							kind: tkSQLStmt,
							str:  " AND dept_no = ",
						},
					},
					Right: &tree{
						Kind: ndEnd,
						Left: &tree{
							Kind: ndEndOfProgram,
							Token: &token{
								kind: tkEndOfProgram,
							},
						},
						Token: &token{
							kind: tkEnd,
							str:  "/* END */",
						},
					},
					Token: &token{
						kind:      tkIf,
						str:       "/* IF false */",
						condition: "false",
					},
				},
				Token: &token{
					kind: tkSQLStmt,
					str:  " ",
				},
			},
			Token: &token{
				kind:  tkBind,
				str:   "?/*maxEmpNo*/",
				value: "maxEmpNo",
			},
		},
		Token: &token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < ",
		},
	}

	wantIfElifElse = &tree{
		Kind: ndSQLStmt,
		Left: &tree{
			Kind: ndIf,
			Left: &tree{
				Kind: ndSQLStmt,
				Token: &token{
					kind: tkSQLStmt,
					str:  "AND dept_no =1",
				},
			},
			Right: &tree{
				Kind: ndElif,
				Left: &tree{
					Kind: ndSQLStmt,
					Token: &token{
						kind: tkSQLStmt,
						str:  " AND boss_no = 2 ",
					},
				},
				Right: &tree{
					Kind: ndElse,
					Left: &tree{
						Kind: ndSQLStmt,
						Token: &token{
							kind: tkSQLStmt,
							str:  " AND id=3",
						},
					},
					Right: &tree{
						Kind: ndEnd,
						Left: &tree{
							Kind: ndEndOfProgram,
							Token: &token{
								kind: tkEndOfProgram,
							},
						},
						Token: &token{
							kind: tkEnd,
							str:  "/* END */",
						},
					},
					Token: &token{
						kind: tkElse,
						str:  "/*ELSE */",
					},
				},
				Token: &token{
					kind:      tkElif,
					str:       "/* ELIF true*/",
					condition: "true",
				},
			},
			Token: &token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	wantIfNest = &tree{
		Kind: ndSQLStmt,
		Left: &tree{
			Kind: ndIf,
			Left: &tree{
				Kind: ndSQLStmt,
				Left: &tree{
					Kind: ndIf,
					Left: &tree{
						Kind: ndSQLStmt,
						Token: &token{
							kind: tkSQLStmt,
							str:  " AND dept_no =1 ",
						},
					},
					Right: &tree{
						Kind: ndElse,
						Left: &tree{
							Kind: ndSQLStmt,
							Token: &token{
								kind: tkSQLStmt,
								str:  " AND id=3 ",
							},
						},
						Right: &tree{
							Kind: ndEnd,
							Left: &tree{
								Kind: ndSQLStmt,
								Token: &token{
									kind: tkSQLStmt,
									str:  " ",
								},
							},
							Token: &token{
								kind: tkEnd,
								str:  "/* END */",
							},
						},
						Token: &token{
							kind: tkElse,
							str:  "/* ELSE */",
						},
					},
					Token: &token{
						kind:      tkIf,
						str:       "/* IF false */",
						condition: "false",
					},
				},
				Token: &token{
					kind: tkSQLStmt,
					str:  " ",
				},
			},
			Right: &tree{
				Kind: ndElse,
				Left: &tree{
					Kind: ndSQLStmt,
					Token: &token{
						kind: tkSQLStmt,
						str:  " AND boss_id=4 ",
					},
				},
				Right: &tree{
					Kind: ndEnd,
					Left: &tree{
						Kind: ndEndOfProgram,
						Token: &token{
							kind: tkEndOfProgram,
						},
					},
					Token: &token{
						kind: tkEnd,
						str:  "/* END */",
					},
				},
				Token: &token{
					kind: tkElse,
					str:  "/* ELSE*/",
				},
			},
			Token: &token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
)
