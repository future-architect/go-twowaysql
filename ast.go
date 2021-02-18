package twowaysql

import (
	"fmt"
)

type NodeKind int

const (
	NdSQLstmt NodeKind = iota + 1
	NdBind
	NdIf
	NdElif
	NdElse
	NdEnd
	NdEndOfProgram
)

type Tree struct {
	Kind  NodeKind
	Left  *Tree
	Right *Tree
	Token *Token
}

// astはトークン列から抽象構文木を生成する。
// 生成規則: （自信はない)
// program = stmt
// stmt = 	SQLStmt stmt |
//			BIND	stmt |
//			EndOfProgram |
//		  	"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END" stmt
//
func ast(tokens []Token) (*Tree, error) {
	node, err := program(tokens)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func program(tokens []Token) (*Tree, error) {
	index := 0
	var node *Tree
	var err error

	node, err = stmt(tokens, &index)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// token index token[index]を見ている
func stmt(tokens []Token, index *int) (*Tree, error) {
	var node *Tree
	var err error
	if consume(tokens, index, TkSQLStmt) {
		node = &Tree{
			Kind:  NdSQLstmt,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}

	} else if consume(tokens, index, TkBind) {
		node = &Tree{
			Kind:  NdBind,
			Token: &tokens[*index-1],
		}

		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
	} else if consume(tokens, index, TkEndOfProgram) {
		node = &Tree{
			Kind:  NodeKind(TkEndOfProgram),
			Token: &tokens[*index-1],
		}
		return node, nil
	} else if consume(tokens, index, TkIf) {
		node = &Tree{
			Kind:  NdIf,
			Token: &tokens[*index-1],
		}
		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		tmpNode := node
		for {
			if consume(tokens, index, TkElif) {
				child := &Tree{
					Kind:  NdElif,
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
		if consume(tokens, index, TkElse) {
			child := &Tree{
				Kind:  NdElse,
				Token: &tokens[*index-1],
			}
			tmpNode.Right = child
			tmpNode = child

			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		}

		if consume(tokens, index, TkEnd) {
			child := &Tree{
				Kind:  NdEnd,
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

		return node, nil
	}
	return node, nil
}

func consume(tokens []Token, index *int, kind TokenKind) bool {
	if tokens[*index].kind == kind {
		*index++
		return true
	}
	return false
}
