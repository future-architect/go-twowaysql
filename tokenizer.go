package twowaysql

import (
	"errors"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tkSQLStmt tokenKind = iota + 1
	tkIf
	tkElif
	tkElse
	tkEnd
	tkBind
	tkEndOfProgram
)

// token holds the information of the token
type token struct {
	kind      tokenKind
	str       string
	value     string /* for Bind */
	condition string /* for IF/ELIF */
}

//tokenizeは文字列を受け取ってトークンの列を返す。
//Token構造体をどのように設計するのが良いかはまだよく分からない
func tokinize(str string) ([]token, error) {
	var tokens []token

	index := 0
	start := 0
	length := len(str)
	//index out of boundsを避けるため末尾に空白を追加する。
	str = str + "    "

	for index < length {
		if str[index:index+2] == "/*" {
			//コメントの直前の塊をTKSQLStmtとしてappend
			tokens = append(tokens, token{
				kind: tkSQLStmt,
				str:  str[start:index],
			})
			start = index
			index += 2
			tok := token{}
			for index < length && str[index:index+2] != "*/" {
				if str[index:index+2] == "IF" {
					tok.kind = tkIf
					index += 2
					continue
				}
				if str[index:index+4] == "ELIF" {
					tok.kind = tkElif
					index += 4
					continue
				}
				if str[index:index+4] == "ELSE" {
					tok.kind = tkElse
					index += 4
					continue
				}
				if str[index:index+3] == "END" {
					tok.kind = tkEnd
					index += 3
					continue
				}
				index++
			}
			// */がなければ不正なフォーマット
			if str[index:index+2] != "*/" {
				return []token{}, errors.New("Comment enclosing characters do not match")
			}
			index += 2
			if tok.kind == 0 {
				tok.kind = tkBind
				if quote := str[index : index+1]; quote == `'` || quote == `"` {
					// 文字列が続いている。
					// 実装汚い...
					index++
					for index < length && str[index:index+1] != quote {
						index++
					}
					if str[index:index+1] != quote {
						return nil, errors.New("Enclosing characters do not match")
					}
					index++
				} else {
					for index < length && str[index] != ' ' {
						index++
					}
				}
			}

			tok.str = str[start:index]
			switch tok.kind {
			case tkIf:
				tok.condition = retrieveConditionFromIf(tok.str)
			case tkElif:
				tok.condition = retrieveConditionFromElif(tok.str)
			case tkBind:
				tok.str = bindLiteral(tok.str)
				tok.value = retrieveValue(tok.str)
			}
			start = index
			tokens = append(tokens, tok)
		}
		if index == length-1 {
			tokens = append(tokens, token{
				kind: tkSQLStmt,
				str:  str[start : index+1],
			})
		}
		index++
	}

	//処理しやすいように終点Tokenを付与する
	tokens = append(tokens, token{
		kind: tkEndOfProgram,
	})
	return tokens, nil
}

// ?/*value*/から value1を取り出す
func retrieveValue(str string) string {
	var retStr string
	retStr = strings.Trim(str, " ")
	retStr = strings.TrimLeft(retStr, "?")
	retStr = strings.TrimPrefix(retStr, "/*")
	retStr = strings.TrimSuffix(retStr, "*/")
	return strings.Trim(retStr, " ")
}

// /*value*/1000 -> ?/*value*/ みたいに変換する
func bindLiteral(str string) string {
	str = strings.TrimRightFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	str = "?" + str
	return str
}

// /* IF condition */ -> condtionを返す
func retrieveConditionFromIf(str string) string {
	str = removeCommentSymbol(str)
	str = strings.Trim(str, " ")
	str = strings.TrimPrefix(str, "IF")
	str = strings.TrimLeft(str, " ")
	return str
}

// /* ELIF condition */ -> condtionを返す
func retrieveConditionFromElif(str string) string {
	str = removeCommentSymbol(str)
	str = strings.Trim(str, " ")
	str = strings.TrimPrefix(str, "ELIF")
	str = strings.TrimLeft(str, " ")
	return str
}

// input: /*value*/ -> output: value
func removeCommentSymbol(str string) string {
	str = strings.TrimPrefix(str, "/*")
	str = strings.TrimSuffix(str, "*/")
	return str
}
