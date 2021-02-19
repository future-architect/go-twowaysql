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
					str:  "/* IF false */",
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
	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &NdEndOfProgram1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000  AND dept_no = 1",
		},
	}
	return &NdSQLStmt1

}

func makeTreeIf() *Tree {

	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}

	NdEnd1 := Tree{
		Kind: NdEnd,
		Left: &NdEndOfProgram1,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}

	NdSQLstmt2 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND dept_no = 1",
		},
	}

	NdIf1 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLstmt2,
		Right: &NdEnd1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF true */",
		},
	}

	NdSQLstmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	return &NdSQLstmt1
}

func makeTreeIfBind() *Tree {
	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}
	NdEnd1 := Tree{
		Kind: NdEnd,
		Left: &NdEndOfProgram1,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}
	NdSQLStmt4 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " ",
		},
	}
	NdBind2 := Tree{
		Kind: NdBind,
		Left: &NdSQLStmt4,
		Token: &Token{
			kind: TkBind,
			str:  "/*deptNo*/0",
		},
	}
	NdSQLStmt3 := Tree{
		Kind: NdSQLStmt,
		Left: &NdBind2,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND dept_no = ",
		},
	}
	NdIf1 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLStmt3,
		Right: &NdEnd1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF false */",
		},
	}
	NdSQLStmt2 := Tree{
		Kind: NdSQLStmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " ",
		},
	}
	NdBind1 := Tree{
		Kind: NdBind,
		Left: &NdSQLStmt2,
		Token: &Token{
			kind: TkBind,
			str:  "/*maxEmpNo*/999",
		},
	}
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &NdBind1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < ",
		},
	}
	return &NdSQLStmt1
}

func makeIfElifElse() *Tree {
	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}
	NdEnd1 := Tree{
		Kind: NdEnd,
		Left: &NdEndOfProgram1,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}
	NdSQLStmt4 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND id=3",
		},
	}
	NdElse1 := Tree{
		Kind:  NdElse,
		Left:  &NdSQLStmt4,
		Right: &NdEnd1,
		Token: &Token{
			kind: TkElse,
			str:  "/*ELSE */",
		},
	}
	NdSQLStmt3 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND boss_no = 2 ",
		},
	}
	NdElif1 := Tree{
		Kind:  NdElif,
		Left:  &NdSQLStmt3,
		Right: &NdElse1,
		Token: &Token{
			kind: TkElif,
			str:  "/* ELIF true*/",
		},
	}
	NdSQLStmt2 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "AND dept_no =1",
		},
	}
	NdIf1 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLStmt2,
		Right: &NdElif1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF true */",
		},
	}
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
	return &NdSQLStmt1
}

func makeIfNest() *Tree {
	NdEndOfProgram1 := Tree{
		Kind: NdEndOfProgram,
		Token: &Token{
			kind: TkEndOfProgram,
		},
	}
	NdEnd2 := Tree{
		Kind: NdEnd,
		Left: &NdEndOfProgram1,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}
	NdSQLStmt6 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND boss_id=4 ",
		},
	}
	NdElse2 := Tree{
		Kind:  NdElse,
		Left:  &NdSQLStmt6,
		Right: &NdEnd2,
		Token: &Token{
			kind: TkElse,
			str:  "/* ELSE*/",
		},
	}
	NdSQLStmt5 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " ",
		},
	}
	NdEnd1 := Tree{
		Kind: NdEnd,
		Left: &NdSQLStmt5,
		Token: &Token{
			kind: TkEnd,
			str:  "/* END */",
		},
	}
	NdSQLStmt4 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND id=3 ",
		},
	}
	NdElse1 := Tree{
		Kind:  NdElse,
		Left:  &NdSQLStmt4,
		Right: &NdEnd1,
		Token: &Token{
			kind: TkElse,
			str:  "/* ELSE */",
		},
	}
	NdSQLStmt3 := Tree{
		Kind: NdSQLStmt,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " AND dept_no =1 ",
		},
	}
	NdIf2 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLStmt3,
		Right: &NdElse1,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF false */",
		},
	}
	NdSQLStmt2 := Tree{
		Kind: NdSQLStmt,
		Left: &NdIf2,
		Token: &Token{
			kind: TkSQLStmt,
			str:  " ",
		},
	}
	NdIf1 := Tree{
		Kind:  NdIf,
		Left:  &NdSQLStmt2,
		Right: &NdElse2,
		Token: &Token{
			kind: TkIf,
			str:  "/* IF true */",
		},
	}
	NdSQLStmt1 := Tree{
		Kind: NdSQLStmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}
	return &NdSQLStmt1
}
