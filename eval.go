package twowaysql

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"gitlab.com/osaki-lab/tagscanner/runtimescan"
)

// Eval returns converted query and bind value.
// inputParams takes a tagged struct. Tags must be in the form `map:"tag_name"`.
// The return value is expected to be used to issue queries to the database
func Eval(inputQuery string, inputParams interface{}) (string, []interface{}, error) {
	mapParams := map[string]interface{}{}

	if inputParams != nil {
		err := encode(mapParams, inputParams)
		if err != nil {
			return "", nil, err
		}
	} else {
		mapParams = nil
	}

	tokens, err := tokinize(inputQuery)
	if err != nil {
		return "", nil, err
	}
	tree, err := ast(tokens)
	if err != nil {
		return "", nil, err
	}

	generatedTokens, err := tree.parse(mapParams)
	if err != nil {
		return "", nil, err
	}

	query, params, err := build(generatedTokens, mapParams)
	if err != nil {
		return "", nil, err
	}

	return arrageWhiteSpace(query), params, nil
}

func build(tokens []token, inputParams map[string]interface{}) (string, []interface{}, error) {
	var b strings.Builder
	var params []interface{}
	var err error

	for _, token := range tokens {
		if token.kind == tkBind {
			if elem, ok := inputParams[token.value]; ok {
				switch slice := elem.(type) {
				case []string:
					token.str = bindLiterals(token.str, len(slice))
					if err != nil {
						return "", nil, err
					}
					for _, value := range slice {
						params = append(params, value)
					}
				case []int:
					token.str = bindLiterals(token.str, len(slice))
					if err != nil {
						return "", nil, err
					}
					for _, value := range slice {
						params = append(params, value)
					}
				default:
					params = append(params, elem)
				}
			} else {
				return "", nil, fmt.Errorf("no parameter that matches the bind value: %s", token.value)
			}
		}
		b.WriteString(token.str)
	}
	return b.String(), params, nil
}

// ?/* ... */ -> (?, ?, ?)/* ... */みたいにする
func bindLiterals(str string, number int) string {
	str = strings.TrimLeftFunc(str, func(r rune) bool {
		return r != unicode.SimpleFold('/')
	})
	var b strings.Builder
	b.WriteRune('(')
	for i := 0; i < number; i++ {
		b.WriteRune('?')
		if i != number-1 {
			b.WriteString(", ")
		}
	}
	b.WriteRune(')')

	return fmt.Sprint(b.String(), str)
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
	return strings.Trim(ret, " ")
}

type encoder struct {
	dest map[string]interface{}
}

func (m encoder) ParseTag(name, tagStr, pathStr string, elemType reflect.Type) (tag interface{}, err error) {
	return runtimescan.BasicParseTag(name, tagStr, pathStr, elemType)
}

func (m *encoder) VisitField(tag, value interface{}) (err error) {
	t := tag.(*runtimescan.BasicTag)
	m.dest[t.Tag] = value
	return nil
}

func (m encoder) EnterChild(tag interface{}) (err error) {
	return nil
}

func (m encoder) LeaveChild(tag interface{}) (err error) {
	return nil
}

func encode(dest map[string]interface{}, src interface{}) error {
	enc := &encoder{
		dest: dest,
	}
	return runtimescan.Encode(src, "map", enc)
}
