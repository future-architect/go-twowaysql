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

type token struct {
	kind      tokenKind
	str       string
	value     string /* for Bind */
	condition string /* for IF/ELIF */
	no        int
}

// tokenizeは文字列を受け取ってトークンの列を返す
func tokenize(str string) ([]token, error) {
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
				if quote := str[index]; quote == '(' {
					// /* ... */( ... )
					index++
					for index < length && str[index] != ')' {
						index++
					}
					if str[index] != ')' {
						return nil, errors.New("Enclosing characters do not match")
					}
					index++
				} else if quote := str[index]; quote == '\'' || quote == '"' {
					// /* ... */"..."
					// /* ... */'...'
					// 文字列が続いている。
					// 実装汚い...
					index++
					for index < length && str[index] != quote {
						index++
					}
					if str[index] != quote {
						return nil, errors.New("Enclosing characters do not match")
					}
					index++
				} else {
					for index < length && str[index] != ' ' && str[index] != ',' && str[index] != ')' {
						index++
					}
				}
			}

			tok.str = str[start:index]
			switch tok.kind {
			case tkIf, tkElif:
				tok.condition = retrieveCondition(tok.kind, tok.str)
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

	// 処理しやすいように終点Tokenを付与する
	tokens = append(tokens, token{
		kind: tkEndOfProgram,
	})

	for i := range tokens {
		tokens[i].no = i
	}
	return tokens, nil
}

// ?/*value*/から value1を取り出す
func retrieveValue(str string) string {
	retStr := strings.Trim(str, " ")
	retStr = strings.TrimLeft(retStr, "?")
	retStr = removeCommentSymbol(retStr)
	return strings.Trim(retStr, " ")
}

// /*value*/1000 -> ?/*value*/ みたいに変換する
func bindLiteral(str string) string {
	str = strings.TrimRightFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	return "?" + str
}

// /* (IF|ELIF) condition */ -> conditionを返す
// kind must be tkIf or tkElif
func retrieveCondition(kind tokenKind, str string) string {
	str = removeCommentSymbol(str)
	str = strings.Trim(str, " ")
	switch kind {
	case tkIf:
		str = strings.TrimPrefix(str, "IF")
	case tkElif:
		str = strings.TrimPrefix(str, "ELIF")
	default:
		panic("kind must be tKIF or tkElif")
	}
	return strings.TrimLeft(str, " ")
}

// input: /*value*/ -> output: value
func removeCommentSymbol(str string) string {
	str = strings.TrimPrefix(str, "/*")
	return strings.TrimSuffix(str, "*/")
}
