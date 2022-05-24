package twowaysql

import "fmt"

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

		iftokenGroup, err := parseIftokens(tokens, &idx, mapParams)
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

func parseIftokens(tokens []token, idx *int, mapParams map[string]interface{}) ([]token, error) {
	tmp := []token{}
	iftokenGroup := []token{tokens[*idx]} // IF
	*idx++
	for {
		if *idx >= len(tokens) {
			return nil, fmt.Errorf("can not parse: not found /* END */")
		}
		// nest IF
		if tokens[*idx].kind == tkIf {
			nestTokens, err := parseIftokens(tokens, idx, mapParams)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, nestTokens...)
			// idx は parseIftokens 内で進んでいるためプラスしない
			continue
		}
		// ELSE/ELIF
		if tokens[*idx].kind == tkElse || tokens[*idx].kind == tkElif {
			nestTokens, err := parseCondition(tmp, mapParams)
			if err != nil {
				return nil, err
			}
			// ネストしたブロックを解析する際に末尾に EndOfProgram が付与されるため除去
			if len(nestTokens) > 0 && nestTokens[len(nestTokens)-1].kind == tkEndOfProgram {
				nestTokens = nestTokens[0 : len(nestTokens)-1]
			}
			tmp = nestTokens
			tmp = append(tmp, tokens[*idx]) // ELSE/ELIF を追加
			iftokenGroup = append(iftokenGroup, tmp...)
			tmp = []token{}
			*idx++
			continue
		}

		if tokens[*idx].kind != tkEnd {
			tmp = append(tmp, tokens[*idx])
			*idx++
			continue
		}

		// END
		iftokenGroup = append(iftokenGroup, tmp...)       // IF ブロック内
		iftokenGroup = append(iftokenGroup, tokens[*idx]) // END
		tmp = []token{}
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
