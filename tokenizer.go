package twowaysql

import (
	"errors"
	"strings"
	"unicode"
)

type TokenKind int

const (
	TkSQLStmt TokenKind = iota + 1
	TkIf
	TkElif
	TkElse
	TkEnd
	TkBind
	TkEndOfProgram
)

type Token struct {
	kind      TokenKind
	str       string
	value     string /* for Bind */
	condition string /* for IF/ELIF */
}

//tokenizeは文字列を受け取ってトークンの列を返す。
//Token構造体をどのように設計するのが良いかはまだよく分からない
func tokinize(str string) ([]Token, error) {
	var tokens []Token

	index := 0
	start := 0
	length := len(str)
	//index out of boundsを避けるため末尾に空白を追加する。
	str = str + "    "

	for index < length {
		if str[index:index+2] == "/*" {
			//コメントの直前の塊をTKSQLStmtとしてappend
			tokens = append(tokens, Token{
				kind: TkSQLStmt,
				str:  str[start:index],
			})
			start = index
			index += 2
			token := Token{}
			for index < length && str[index:index+2] != "*/" {
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
				if str[index:index+3] == "END" {
					token.kind = TkEnd
					index += 3
					continue
				}
				index++
			}
			// */がなければ不正なフォーマット
			if str[index:index+2] != "*/" {
				return []Token{}, errors.New("can not tokenize")
			}
			index += 2
			if token.kind == 0 {
				token.kind = TkBind
				for index < length && str[index] != ' ' {
					index++
				}
			}

			token.str = str[start:index]
			switch token.kind {
			case TkIf:
				token.condition = retrieveValueFromIf(token.str)
			case TkElif:
				token.condition = retrieveValueFromElif(token.str)
			case TkBind:
				token.str = bindConvert(token.str)
				token.value = retrieveValue(token.str)
			}
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

	//処理しやすいように終点Tokenを付与する
	tokens = append(tokens, Token{
		kind: TkEndOfProgram,
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
func bindConvert(str string) string {
	str = strings.TrimRightFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	str = "?" + str
	return str
}

// /* IF condition */ -> condtionを返す
func retrieveValueFromIf(str string) string {
	str = removeCommentSymbol(str)
	str = strings.Trim(str, " ")
	str = strings.TrimPrefix(str, "IF")
	str = strings.TrimLeft(str, " ")
	return str
}

// /* ELIF condition */ -> condtionを返す
func retrieveValueFromElif(str string) string {
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
