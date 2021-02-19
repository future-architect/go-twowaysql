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
					kind: TkEndOfProgram,
				},
			},
			want: makeEmpty(),
		},
		{
			name: "no comment",
			input: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
				},
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeNoComment(),
		},
		{
			name: "if",
			input: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = 1",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeTreeIf(),
		},
		{
			name: "if and bind",
			input: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < ",
				},
				{
					kind: TkBind,
					str:  "/*maxEmpNo*/999",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkIf,
					str:  "/* IF exists(deptNo)*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no = ",
				},
				{
					kind: TkBind,
					str:  "/*deptNo*/0",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeTreeIfBind(),
		},
		{
			name: "if elif else",
			input: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  "AND dept_no =1",
				},
				{
					kind: TkElif,
					str:  "/* ELIF true*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND boss_no = 2 ",
				},
				{
					kind: TkElse,
					str:  "/*ELSE */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND id=3",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeIfElifElse(),
		},
		{
			name: "if nest",
			input: []Token{
				{
					kind: TkSQLStmt,
					str:  "SELECT * FROM person WHERE employee_no < 1000 ",
				},
				{
					kind: TkIf,
					str:  "/* IF true */",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkIf,
					str:  "/* IF false */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND dept_no =1 ",
				},
				{
					kind: TkElse,
					str:  "/* ELSE */",
				},
				{
					kind: TkSQLStmt,
					str:  " AND id=3 ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkSQLStmt,
					str:  " ",
				},
				{
					kind: TkElse,
					str:  "/* ELSE*/",
				},
				{
					kind: TkSQLStmt,
					str:  " AND boss_id=4 ",
				},
				{
					kind: TkEnd,
					str:  "/* END */",
				},
				{
					kind: TkEndOfProgram,
				},
			},
			want: makeIfNest(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ast(tt.input); err != nil || !treeEqual(tt.want, got) {
				if err != nil {
					t.Error(err)
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
	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}
	return &NdEndOfProgram1

}

func makeNoComment() *Tree {
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &Tree{
			Kind: NdEndOfProgram,
			Token: &Token{
				kind: TkEndOfProgram,
			},
		},
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
		},
	}
	return &NdSQLStmt1

}

func makeTreeIf() *Tree {

	NdSQLstmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &Tree{
			Kind: NdIf,
			Left: &Tree{
				Kind: NdSQLStmt,
				Token: &Token{
					kind: TkSQLStmt,
					str:  " AND dept_no = 1",
				},
			},
			Right: &Tree{
				Kind: NdEnd,
				Left: &Tree{
					Kind: NdEndOfProgram,
					Token: &Token{
						kind: TkEndOfProgram,
					},
				},
				Token: &Token{
					kind: TkEnd,
					str:  "/* END */",
				},
			},
			Token: &Token{
				kind: TkIf,
				str:  "/* IF true */",
			},
		},
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	return &NdSQLstmt1
}

func makeTreeIfBind() *Tree {
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &Tree{
			Kind: NdBind,
			Left: &Tree{
				Kind: NdSQLStmt,
				Left: &Tree{
					Kind: NdIf,
					Left: &Tree{
						Kind: NdSQLStmt,
						Left: &Tree{
							Kind: NdBind,
							Left: &Tree{
								Kind: NdSQLStmt,
								Token: &Token{
									kind: TkSQLStmt,
									str:  " ",
								},
							},
							Token: &Token{
								kind: TkBind,
								str:  "/*deptNo*/0",
							},
						},
						Token: &Token{
							kind: TkSQLStmt,
							str:  " AND dept_no = ",
						},
					},
					Right: &Tree{
						Kind: NdEnd,
						Left: &Tree{
							Kind: NdEndOfProgram,
							Token: &Token{
								kind: TkEndOfProgram,
							},
						},
						Token: &Token{
							kind: TkEnd,
							str:  "/* END */",
						},
					},
					Token: &Token{
						kind: TkIf,
						str:  "/* IF exists(deptNo)*/",
					},
				},
				Token: &Token{
					kind: TkSQLStmt,
					str:  " ",
				},
			},
			Token: &Token{
				kind: TkBind,
				str:  "/*maxEmpNo*/999",
			},
		},
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < ",
		},
	}
	return &NdSQLStmt1
}

func makeIfElifElse() *Tree {
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &Tree{
			Kind: NdIf,
			Left: &Tree{
				Kind: NdSQLStmt,
				Token: &Token{
					kind: TkSQLStmt,
					str:  "AND dept_no =1",
				},
			},
			Right: &Tree{
				Kind: NdElif,
				Left: &Tree{
					Kind: NdSQLStmt,
					Token: &Token{
						kind: TkSQLStmt,
						str:  " AND boss_no = 2 ",
					},
				},
				Right: &Tree{
					Kind: NdElse,
					Left: &Tree{
						Kind: NdSQLStmt,
						Token: &Token{
							kind: TkSQLStmt,
							str:  " AND id=3",
						},
					},
					Right: &Tree{
						Kind: NdEnd,
						Left: &Tree{
							Kind: NdEndOfProgram,
							Token: &Token{
								kind: TkEndOfProgram,
							},
						},
						Token: &Token{
							kind: TkEnd,
							str:  "/* END */",
						},
					},
					Token: &Token{
						kind: TkElse,
						str:  "/*ELSE */",
					},
				},
				Token: &Token{
					kind: TkElif,
					str:  "/* ELIF true*/",
				},
			},
			Token: &Token{
				kind: TkIf,
				str:  "/* IF true */",
			},
		},
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
	return &NdSQLStmt1
}

func makeIfNest() *Tree {
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &Tree{
			Kind: NdIf,
			Left: &Tree{
				Kind: NdSQLStmt,
				Left: &Tree{
					Kind: NdIf,
					Left: &Tree{
						Kind: NdSQLStmt,
						Token: &Token{
							kind: TkSQLStmt,
							str:  " AND dept_no =1 ",
						},
					},
					Right: &Tree{
						Kind: NdElse,
						Left: &Tree{
							Kind: NdSQLStmt,
							Token: &Token{
								kind: TkSQLStmt,
								str:  " AND id=3 ",
							},
						},
						Right: &Tree{
							Kind: NdEnd,
							Left: &Tree{
								Kind: NdSQLStmt,
								Token: &Token{
									kind: TkSQLStmt,
									str:  " ",
								},
							},
							Token: &Token{
								kind: TkEnd,
								str:  "/* END */",
							},
						},
						Token: &Token{
							kind: TkElse,
							str:  "/* ELSE */",
						},
					},
					Token: &Token{
						kind: TkIf,
						str:  "/* IF false */",
					},
				},
				Token: &Token{
					kind: TkSQLStmt,
					str:  " ",
				},
			},
			Right: &Tree{
				Kind: NdElse,
				Left: &Tree{
					Kind: NdSQLStmt,
					Token: &Token{
						kind: TkSQLStmt,
						str:  " AND boss_id=4 ",
					},
				},
				Right: &Tree{
					Kind: NdEnd,
					Left: &Tree{
						Kind: NdEndOfProgram,
						Token: &Token{
							kind: TkEndOfProgram,
						},
					},
					Token: &Token{
						kind: TkEnd,
						str:  "/* END */",
					},
				},
				Token: &Token{
					kind: TkElse,
					str:  "/* ELSE*/",
				},
			},
			Token: &Token{
				kind: TkIf,
				str:  "/* IF true */",
			},
		},
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
	return &NdSQLStmt1
}
