package twowaysql

import (
	"fmt"
	"testing"
)

//　課題: テストケースの準備(Treeの作成)に手間がかかる
func TestAst(t *testing.T) {
	tests := []struct {
		name  string
		input []Token
		want  *Tree
	}{
		{
			name: "empty",
			input: []Token{
				{
					kind: tkEndOfProgram,
				},
			},
			want: makeEmpty(),
		},
		{
			name: "no comment",
			input: []Token{
				{
					kind: tkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
				{
					kind: tkEndOfProgram,
				},
			},
			want: makeNoComment(),
		},
		{
			name: "if",
			input: []Token{
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
			want: makeTreeIf(),
		},
		{
			name: "if and bind",
			input: []Token{
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
			want: makeTreeIfBind(),
		},
		{
			name: "if elif else",
			input: []Token{
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
			want: makeIfElifElse(),
		},
		{
			name: "if nest",
			input: []Token{
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
			want: makeIfNest(),
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

func walk(t *Tree, ch chan *Token) {
	defer close(ch)

	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)

}

func walkInner(t *Tree, ch chan *Token) {
	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)
}

func treeEqual(t1, t2 *Tree) bool {
	ch1 := make(chan *Token)
	ch2 := make(chan *Token)

	go walk(t1, ch1)
	go walk(t2, ch2)

	var s1, s2 []*Token

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

func printTree(t *Tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)

}

func printWalkInner(t *Tree) {
	if t == nil {
		return
	}

	printWalkInner(t.Left)
	fmt.Println(t.Token)
	printWalkInner(t.Right)
}

// テストの期待する結果を作成
func makeEmpty() *Tree {
	return &Tree{
		Kind: ndEndOfProgram,
		Token: &Token{
			kind: tkEndOfProgram,
		},
	}
}

func makeNoComment() *Tree {
	return &Tree{
		Kind: ndSQLStmt,
		Left: &Tree{
			Kind: ndEndOfProgram,
			Token: &Token{
				kind: tkEndOfProgram,
			},
		},
		Token: &Token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
		},
	}
}

func makeTreeIf() *Tree {
	return &Tree{
		Kind: ndSQLStmt,
		Left: &Tree{
			Kind: ndIf,
			Left: &Tree{
				Kind: ndSQLStmt,
				Token: &Token{
					kind: tkSQLStmt,
					str:  " AND dept_no = 1",
				},
			},
			Right: &Tree{
				Kind: ndEnd,
				Left: &Tree{
					Kind: ndEndOfProgram,
					Token: &Token{
						kind: tkEndOfProgram,
					},
				},
				Token: &Token{
					kind: tkEnd,
					str:  "/* END */",
				},
			},
			Token: &Token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &Token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
}

func makeTreeIfBind() *Tree {
	return &Tree{
		Kind: ndSQLStmt,
		Left: &Tree{
			Kind: ndBind,
			Left: &Tree{
				Kind: ndSQLStmt,
				Left: &Tree{
					Kind: ndIf,
					Left: &Tree{
						Kind: ndSQLStmt,
						Left: &Tree{
							Kind: ndBind,
							Left: &Tree{
								Kind: ndSQLStmt,
								Token: &Token{
									kind: tkSQLStmt,
									str:  " ",
								},
							},
							Token: &Token{
								kind:  tkBind,
								str:   "?/*deptNo*/",
								value: "deptNo",
							},
						},
						Token: &Token{
							kind: tkSQLStmt,
							str:  " AND dept_no = ",
						},
					},
					Right: &Tree{
						Kind: ndEnd,
						Left: &Tree{
							Kind: ndEndOfProgram,
							Token: &Token{
								kind: tkEndOfProgram,
							},
						},
						Token: &Token{
							kind: tkEnd,
							str:  "/* END */",
						},
					},
					Token: &Token{
						kind:      tkIf,
						str:       "/* IF false */",
						condition: "false",
					},
				},
				Token: &Token{
					kind: tkSQLStmt,
					str:  " ",
				},
			},
			Token: &Token{
				kind:  tkBind,
				str:   "?/*maxEmpNo*/",
				value: "maxEmpNo",
			},
		},
		Token: &Token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < ",
		},
	}
}

func makeIfElifElse() *Tree {
	return &Tree{
		Kind: ndSQLStmt,
		Left: &Tree{
			Kind: ndIf,
			Left: &Tree{
				Kind: ndSQLStmt,
				Token: &Token{
					kind: tkSQLStmt,
					str:  "AND dept_no =1",
				},
			},
			Right: &Tree{
				Kind: ndElif,
				Left: &Tree{
					Kind: ndSQLStmt,
					Token: &Token{
						kind: tkSQLStmt,
						str:  " AND boss_no = 2 ",
					},
				},
				Right: &Tree{
					Kind: ndElse,
					Left: &Tree{
						Kind: ndSQLStmt,
						Token: &Token{
							kind: tkSQLStmt,
							str:  " AND id=3",
						},
					},
					Right: &Tree{
						Kind: ndEnd,
						Left: &Tree{
							Kind: ndEndOfProgram,
							Token: &Token{
								kind: tkEndOfProgram,
							},
						},
						Token: &Token{
							kind: tkEnd,
							str:  "/* END */",
						},
					},
					Token: &Token{
						kind: tkElse,
						str:  "/*ELSE */",
					},
				},
				Token: &Token{
					kind:      tkElif,
					str:       "/* ELIF true*/",
					condition: "true",
				},
			},
			Token: &Token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &Token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
}

func makeIfNest() *Tree {
	return &Tree{
		Kind: ndSQLStmt,
		Left: &Tree{
			Kind: ndIf,
			Left: &Tree{
				Kind: ndSQLStmt,
				Left: &Tree{
					Kind: ndIf,
					Left: &Tree{
						Kind: ndSQLStmt,
						Token: &Token{
							kind: tkSQLStmt,
							str:  " AND dept_no =1 ",
						},
					},
					Right: &Tree{
						Kind: ndElse,
						Left: &Tree{
							Kind: ndSQLStmt,
							Token: &Token{
								kind: tkSQLStmt,
								str:  " AND id=3 ",
							},
						},
						Right: &Tree{
							Kind: ndEnd,
							Left: &Tree{
								Kind: ndSQLStmt,
								Token: &Token{
									kind: tkSQLStmt,
									str:  " ",
								},
							},
							Token: &Token{
								kind: tkEnd,
								str:  "/* END */",
							},
						},
						Token: &Token{
							kind: tkElse,
							str:  "/* ELSE */",
						},
					},
					Token: &Token{
						kind:      tkIf,
						str:       "/* IF false */",
						condition: "false",
					},
				},
				Token: &Token{
					kind: tkSQLStmt,
					str:  " ",
				},
			},
			Right: &Tree{
				Kind: ndElse,
				Left: &Tree{
					Kind: ndSQLStmt,
					Token: &Token{
						kind: tkSQLStmt,
						str:  " AND boss_id=4 ",
					},
				},
				Right: &Tree{
					Kind: ndEnd,
					Left: &Tree{
						Kind: ndEndOfProgram,
						Token: &Token{
							kind: tkEndOfProgram,
						},
					},
					Token: &Token{
						kind: tkEnd,
						str:  "/* END */",
					},
				},
				Token: &Token{
					kind: tkElse,
					str:  "/* ELSE*/",
				},
			},
			Token: &Token{
				kind:      tkIf,
				str:       "/* IF true */",
				condition: "true",
			},
		},
		Token: &Token{
			kind: tkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
}
