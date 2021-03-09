package twowaysql

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type state struct {
	query  string
	params []interface{}
}

func Parse(inputQuery string, inputParams map[string]interface{}) (state, error) {
	tokens, err := tokinize(inputQuery)
	if err != nil {
		return state{}, err
	}
	tree, err := ast(tokens)
	if err != nil {
		return state{}, err
	}

	generatedTokens, err := gen(tree, inputParams)
	if err != nil {
		return state{}, err
	}

	query, params, err := build(generatedTokens, inputParams)
	if err != nil {
		return state{}, err
	}

	return state{
		query:  arrageWhiteSpace(query),
		params: params,
	}, nil
}

// TkBindのvalueをtoken列から取り出す
func retrieveBinds(tokens []token) []string {
	var binds []string
	for _, token := range tokens {
		if token.kind == tkBind {
			binds = append(binds, token.value)
		}
	}
	return binds
}

func buildQuery(tokens []token) string {
	var b strings.Builder
	for _, token := range tokens {
		b.WriteString(token.str)
	}

	return arrageWhiteSpace(b.String())
}

func build(tokens []token, inputParams map[string]interface{}) (string, []interface{}, error) {
	var b strings.Builder
	var params []interface{}
	var err error

	for _, token := range tokens {
		if token.kind == tkBind {
			if elem, ok := inputParams[token.value]; ok {
				if slice, ok := elem.([]string); ok {
					token.str, err = bindLiterals(token.str, len(slice))
					if err != nil {
						return "", nil, err
					}
					for _, value := range slice {
						params = append(params, value)
					}
				} else {
					params = append(params, elem)
				}
			} else {
				return "", nil, fmt.Errorf("no parameter that matches the bind value: %s", token.value)
			}
		}
		_, err = b.WriteString(token.str)
		if err != nil {
			return "", nil, err
		}
	}
	return b.String(), params, nil
}

// ?/* ... */ -> (?, ?, ?)/* ... */みたいにする
func bindLiterals(str string, number int) (string, error) {
	str = strings.TrimLeftFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	var b strings.Builder
	_, err := b.WriteRune('(')
	if err != nil {
		return "", err
	}
	for i := 0; i < number; i++ {
		_, err := b.WriteRune('?')
		if err != nil {
			return "", err
		}
		if i != number-1 {
			_, err := b.WriteString(", ")
			if err != nil {
				return "", err
			}
		}
	}
	_, err = b.WriteRune(')')
	if err != nil {
		return "", err
	}

	return fmt.Sprint(b.String(), str), nil
}

// 空白が二つ以上続いていたら一つにする。=1 -> = 1のような変換はできない
// 単純な空白を想定。 -> issue: よりロバストな実装
func arrageWhiteSpace(str string) string {
	ret := ""
	buff := bytes.NewBufferString(ret)
	for i := 0; i < len(str); i++ {
		if i < len(str)-1 && str[i] == ' ' && str[i+1] == ' ' {
			continue
		}
		buff.WriteByte(str[i])
	}
	ret = buff.String()
	ret = strings.TrimLeft(ret, " ")
	return strings.TrimRight(ret, " ")
}
