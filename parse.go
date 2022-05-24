package twowaysql

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type tokenGroup struct {
	tokens []token
}

func parseCondition(tokens []token, mapParams map[string]interface{}) ([]token, error) {
	var tokenGroups []tokenGroup
	var tmpTokens []token
	var idx int
	for idx < len(tokens) {
		if tokens[idx].kind != tkIf {
			tmpTokens = append(tmpTokens, tokens[idx])
			idx++
			continue
		}
		ng := tokenGroup{
			tokens: tmpTokens,
		}
		tokenGroups = append(tokenGroups, ng)
		tmpTokens = []token{}

		iftokenGroup, err := parseIftokenGroup(tokens, &idx, mapParams)
		if err != nil {
			return nil, err
		}
		ng = tokenGroup{
			tokens: iftokenGroup,
		}
		tokenGroups = append(tokenGroups, ng)
	}
	if len(tmpTokens) != 0 {
		tokenGroups = append(tokenGroups, tokenGroup{tokens: tmpTokens})
	}

	var generatedTokens []token
	for _, taskGroup := range tokenGroups {
		generatedTokens = append(generatedTokens, taskGroup.tokens...)
	}

	if generatedTokens[len(generatedTokens)-1].kind != tkEndOfProgram {
		// 末尾に EndOfProgram を追加
		generatedTokens = append(generatedTokens, token{kind: tkEndOfProgram})
	}
	// 構文エラーチェックのため生成 token 全体でも解析を行う
	_, err := ast(generatedTokens)
	if err != nil {
		return nil, err
	}

	return generatedTokens, nil
}

func parseIftokenGroup(tokens []token, idx *int, mapParams map[string]interface{}) ([]token, error) {
	tmpTokens := []token{}
	iftokenGroup := []token{tokens[*idx]} // IF
	*idx++
	for {
		if *idx >= len(tokens) {
			return nil, fmt.Errorf("can not parse: not found /* END */")
		}
		// nest IF
		if tokens[*idx].kind == tkIf {
			nestTokens, err := parseIftokenGroup(tokens, idx, mapParams)
			if err != nil {
				return nil, err
			}
			tmpTokens = append(tmpTokens, nestTokens...)
			// idx は parseIftokens 内で進んでいるためプラスしない
			continue
		}
		// ELSE/ELIF
		if tokens[*idx].kind == tkElse || tokens[*idx].kind == tkElif {
			nestTokens, err := parseCondition(tmpTokens, mapParams)
			if err != nil {
				return nil, err
			}
			// ネストしたブロックを解析する際に末尾に EndOfProgram が付与されるため除去
			if len(nestTokens) > 0 && nestTokens[len(nestTokens)-1].kind == tkEndOfProgram {
				nestTokens = nestTokens[0 : len(nestTokens)-1]
			}
			tmpTokens = nestTokens
			tmpTokens = append(tmpTokens, tokens[*idx]) // ELSE/ELIF を追加
			iftokenGroup = append(iftokenGroup, tmpTokens...)
			tmpTokens = []token{}
			*idx++
			continue
		}

		if tokens[*idx].kind != tkEnd {
			tmpTokens = append(tmpTokens, tokens[*idx])
			*idx++
			continue
		}

		// END
		iftokenGroup = append(iftokenGroup, tmpTokens...)       // IF ブロック内
		iftokenGroup = append(iftokenGroup, tokens[*idx]) // END
		tmpTokens = []token{}
		*idx++
		break
	}

	// IF ブロックのみで解析するため、末尾に EndOfProgram を追加
	iftokenGroup = append(iftokenGroup, token{kind: tkEndOfProgram})
	tree, err := ast(iftokenGroup)
	if err != nil {
		return nil, err
	}
	generatedTokens, err := tree.parse(mapParams)
	if err != nil {
		return nil, err
	}
	if len(generatedTokens) > 0 && generatedTokens[len(generatedTokens)-1].kind == tkEndOfProgram {
		// 末尾の EndOfProgram を除去
		generatedTokens = generatedTokens[0 : len(generatedTokens)-1]
	}

	return generatedTokens, nil
}

// 抽象構文木からトークン列を生成
// 左部分木、右部分木と辿る
// 現状右部分木を持つのはif, elif, elseだけ?
func (t *tree) parse(params map[string]interface{}) ([]token, error) {
	return genInner(t, params)
}

func genInner(node *tree, params map[string]interface{}) ([]token, error) {
	if node == nil {
		return []token{}, nil
	}

	// 行きがけ

	// 左部分木に行く
	leftStr, err := genInner(node.Left, params)
	if err != nil {
		return []token{}, err
	}

	// 左部分木から戻ってきた

	// 右部分木に行く
	rightStr, err := genInner(node.Right, params)
	if err != nil {
		return []token{}, err
	}

	// 右部分木から戻ってきた
	// 何を返すか
	// 基本的に左部分木
	// If Elifの場合は条件次第
	switch kind := node.Kind; kind {
	case ndSQLStmt, ndBind:
		// めちゃめちゃ実行効率悪い気が...
		return append([]token{*node.Token}, leftStr...), nil
	case ndIf, ndElif:
		truth, err := evalCondition(node.Token.condition, params)
		if err != nil {
			return []token{}, err
		}
		if truth {
			return leftStr, nil
		}
		return rightStr, nil
	default:
		return leftStr, nil
	}
}

// /* If ... */ /* Elif ... */の条件を評価する
func evalCondition(condition string, params map[string]interface{}) (bool, error) {
	vm := otto.New()
	for key, value := range params {
		err := vm.Set(key, value)
		if err != nil {
			return false, err
		}
	}

	result, err := vm.Run(condition)
	if err != nil {
		return false, err
	}

	truth, err := result.ToBoolean()
	if err != nil {
		return false, err
	}

	return truth, nil
}
