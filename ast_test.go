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
			want: makeTreeif(),
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
			want: nil,
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
	if t == nil {
		return
	}

	walkInner(t.Left, ch)
	ch <- t.Token
	walkInner(t.Right, ch)

	close(ch)
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
	}
	return true
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

func makeTreeif() *Tree {

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
		Kind: NdSQLstmt,
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
		Kind: NdSQLstmt,
		Left: &NdIf1,
		Token: &Token{
			kind: TkSQLStmt,
			str:  "SELECT * FROM person WHERE employee_no < 1000 ",
		},
	}

	return &NdSQLstmt1
}
