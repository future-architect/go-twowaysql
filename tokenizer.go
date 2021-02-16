package twowaysql

import "errors"

type TokenKind int

const (
	TkSQLStmt TokenKind = iota + 1
	TkIf
	TkElif
	TkElse
	TkBind
)

type Token struct {
	kind TokenKind
	str  string
}

func tokinize(str string) ([]Token, error) {
	var tokens []Token

	index := 0
	start := 0
	length := len(str)
	//index out of boundsを避けるため末尾に空白を追加する。
	str = str + "    "

	for index < length {
		if str[index:index+2] == "/*" {
			tokens = append(tokens, Token{
				kind: TkSQLStmt,
				str:  str[start:index],
			})
			start = index
			index += 2
			token := Token{}
			for index < length-1 && str[index:index+2] != "*/" {
				if str[index:index+2] == "IF" {
					token.kind = TkIf
					index += 2
					continue
				}
				if str[index:index+4] == "ELIF" {
					token.kind = TkElif
					index += 4
					continue
				}
				if str[index:index+4] == "ELSE" {
					token.kind = TkElse
					index += 4
					continue
				}
				index++
			}
			if str[index:index+2] != "*/" {
				return []Token{}, errors.New("can not tokenize")
			}
			if token.kind == 0 {
				token.kind = TkBind
			}
			index += 2
			token.str = str[start:index]
			start = index
			tokens = append(tokens, token)
		}
		if index == length-1 {
			tokens = append(tokens, Token{
				kind: TkSQLStmt,
				str:  str[start : index+1],
			})
		}
		index++
	}
	return tokens, nil
}
