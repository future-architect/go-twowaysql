package twowaysql

import "errors"

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
// program = stmt*
// stmt = 	SQLStmt |
//			BIND	|
//		  	"IF" stmt ("ELLF" stmt)* ("ELSE" stmt)? "END"
//
func ast(tokens []Token) (*Tree, error) {
	node, err := program(tokens)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// channelで送るようにする?
func program(tokens []Token) (*Tree, error) {
	index := 0
	node := &Tree{}
	tmpNode := node
	var err error

	for {
		tmpNode.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		//根元はNdEndOfProgramではない
		if tmpNode.Kind == NdEndOfProgram {
			return node, nil
		}
		tmpNode = tmpNode.Left
	}
}

// token index token[index]を見ている
// indexの操作が非常に煩雑でerror prone
func stmt(tokens []Token, index int) (*Tree, error) {
	var node *Tree
	if check(tokens, index, TkEndOfProgram) {
		//ここでchannel close?
		node = &Tree{
			Kind:  NdEndOfProgram,
			Token: &tokens[index],
		}
		return node, nil
	}
	if check(tokens, index, TkSQLStmt) {
		node = &Tree{
			Kind:  NdSQLstmt,
			Token: &tokens[index],
		}
		return node, nil
	} else if check(tokens, index, TkBind) {
		node = &Tree{
			Kind:  NdBind,
			Token: &tokens[index],
		}
		return node, nil
	} else if check(tokens, index, TkIf) {
		var err error
		node = &Tree{
			Kind:  NdIf,
			Token: &tokens[index],
		}
		index++
		node.Left, err = stmt(tokens, index)
		if err != nil {
			return nil, err
		}
		tmpNode := node
		for {
			index++
			if check(tokens, index, TkElif) {
				child := &Tree{
					Kind:  NdElif,
					Token: &tokens[index],
				}
				tmpNode.Right = child
				tmpNode = child

				index++
				child.Left, err = stmt(tokens, index)
				if err != nil {
					return nil, err
				}
				continue
			}
			break
		}
		if check(tokens, index+1, TkElse) {
			child := &Tree{
				Kind:  NdElse,
				Token: &tokens[index],
			}
			tmpNode.Right = child
			tmpNode = child

			index++
			child.Left, err = stmt(tokens, index)
			if err != nil {
				return nil, err
			}
		}

		if check(tokens, index, TkEnd) {
			index++
			child := &Tree{
				Kind:  NdEnd,
				Token: &tokens[index],
			}
			tmpNode.Right = child
		} else {
			return nil, errors.New("can not parse: expected /* END */, but can not find")
		}
		return node, nil
	} else {
		return nil, errors.New("can not parse")
	}
}

func check(tokens []Token, index int, kind TokenKind) bool {
	return tokens[index].kind == kind
}
