package twowaysql

import (
	"bytes"
	"strings"
)

type state struct {
	tokens     []Token
	bindsValue []string
	query      string
}

func (t *Twowaysql) parse() (state, error) {
	tokens, err := tokinize(t.query)
	if err != nil {
		return state{}, err
	}
	tree, err := ast(tokens)
	if err != nil {
		return state{}, err
	}

	generatedTokens, err := gen(tree, t.params)
	if err != nil {
		return state{}, err
	}

	query := buildQuery(generatedTokens)
	binds := retrieveBinds(generatedTokens)

	return state{
		tokens:     generatedTokens,
		bindsValue: binds,
		query:      query,
	}, nil

}

// TkBindのvalueをtoken列から取り出す
func retrieveBinds(tokens []Token) []string {
	var binds []string
	for _, token := range tokens {
		if token.kind == TkBind {
			binds = append(binds, token.value)
		}
	}
	return binds
}

func buildQuery(tokens []Token) string {
	var b strings.Builder
	for _, token := range tokens {
		b.WriteString(token.str)
	}

	return arrageWhiteSpace(b.String())
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
	ret = strings.TrimRight(ret, " ")
	return ret
}
