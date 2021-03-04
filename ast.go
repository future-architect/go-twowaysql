package twowaysql

import (
	"errors"
	"fmt"
)

type nodeKind int

const (
	ndSQLStmt nodeKind = iota + 1
	ndBind
	ndIf
	ndElif
	ndElse
	ndEnd
	ndEndOfProgram
)

// Tree is a component of an abstract syntax tree
type Tree struct {
	Kind  nodeKind
	Left  *Tree
	Right *Tree
	Token *Token
}

// astはトークン列から抽象構文木を生成する。
// 生成規則: （自信はない)
// program = stmt
// stmt = 	SQLStmt stmt |
//			BIND	stmt |
//		  	"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END" stmt |
//			EndOfProgram
//
func ast(tokens []Token) (*Tree, error) {
	node, err := program(tokens)
	if err != nil {
		return nil, err
	}
	if node.nodeCount() != len(tokens) {
		//log.Println("node", nodeCount(node), "tokens", len(tokens))
		return nil, errors.New("can not generate abstract syntax tree")
	}

	return node, nil
}

func program(tokens []Token) (*Tree, error) {
	index := 0

	node, err := stmt(tokens, &index)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// token index token[index]を見ている
// 課題：不正な形式でもエラーが返らないと思う。ただ正しくない結果が返ってくる
func stmt(tokens []Token, index *int) (*Tree, error) {
	var node *Tree
	var err error
	if consume(tokens, index, tkSQLStmt) {
		// SQLStmt stmt
		node = &Tree{
			Kind:  ndSQLStmt,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}

	} else if consume(tokens, index, tkBind) {
		// Bind stmt
		node = &Tree{
			Kind:  ndBind,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
	} else if consume(tokens, index, tkEndOfProgram) {
		// EndOfProgram
		node = &Tree{
			Kind: nodeKind(tkEndOfProgram),
			// consumeはTkEndOfProgramの時はインクリメントしないから1を引かない
			// かなりよくない設計
			Token: &tokens[*index],
		}
		return node, nil
	} else if consume(tokens, index, tkIf) {
		//"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END" stmt
		node = &Tree{
			Kind:  ndIf,
			Token: &tokens[*index-1],
		}
		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		tmpNode := node
		for {
			//("ELLF" stmt)*
			if consume(tokens, index, tkElif) {
				child := &Tree{
					Kind:  ndElif,
					Token: &tokens[*index-1],
				}
				tmpNode.Right = child
				tmpNode = child

				child.Left, err = stmt(tokens, index)
				if err != nil {
					return nil, err
				}
				continue
			}
			break
		}
		if consume(tokens, index, tkElse) {
			//("ELSE" stmt)?
			child := &Tree{
				Kind:  ndElse,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child
			tmpNode = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		}

		if consume(tokens, index, tkEnd) {
			//"END"
			child := &Tree{
				Kind:  ndEnd,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("can not parse: expected /* END */, but got %v", tokens[*index].kind)
		}

		//どれも一致しなかった
		return node, nil
	}
	return node, nil
}

//tokenが所望のものか調べる。一致していればインデックスを一つ進める
func consume(tokens []Token, index *int, kind tokenKind) bool {
	//println("str: ", tokens[*index].str, "kind: ", tokens[*index].kind, "want kind: ", kind)
	if tokens[*index].kind == kind {
		// TkEndOfPraogramでインクリメントしてしまうと
		// その後のconsume呼び出しでIndex Out Of Bounds例外が発生してしまう
		if kind != tkEndOfProgram {
			*index++
		}
		return true
	}
	return false
}

func (t *Tree) nodeCount() int {
	count := 1
	if t.Left != nil {
		t.Left.countInner(&count)
	}
	if t.Right != nil {
		t.Right.countInner(&count)
	}
	return count
}

func (t *Tree) countInner(count *int) {
	*count++
	if t.Left != nil {
		t.Left.countInner(count)
	}
	if t.Right != nil {
		t.Right.countInner(count)
	}
}
