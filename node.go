package twowaysql

import "fmt"

type nodeGroupTyp int

const (
	ngSQLStmt nodeGroupTyp = iota + 1
	ngIf
	ngEndOfProgram
)

type nodeGroup struct {
	tokens []token
	typ    nodeGroupTyp
	tree   *tree
}

func splitNodeGroup(tokens []token, mapParams map[string]interface{}) ([]token, error) {
	var nodeGroups []nodeGroup
	var tmp []token
	var idx int
	for idx < len(tokens) {
		if tokens[idx].kind != tkIf {
			tmp = append(tmp, tokens[idx])
			idx++
			continue
		}
		tmp = append(tmp, token{kind: tkEndOfProgram}) // 分割ごとに追加
		ng := nodeGroup{
			tokens: tmp,
		}
		nodeGroups = append(nodeGroups, ng)

		iftokenGroup, err := recursiveIfnestt(tokens, &idx, mapParams)
		if err != nil {
			return nil, err
		}
		iftokenGroup = append(iftokenGroup, token{kind: tkEndOfProgram})
		ng = nodeGroup{
			tokens: iftokenGroup,
		}
		nodeGroups = append(nodeGroups, ng)
		tmp = []token{}
	}
	if len(tmp) != 0 {
		if tmp[len(tmp)-1].kind != tkEndOfProgram {
			tmp = append(tmp, token{kind: tkEndOfProgram}) // 分割ごとに追加
		}
		nodeGroups = append(nodeGroups, nodeGroup{tokens: tmp})
	}

	var generatedTokens []token
	for i := range nodeGroups {
		tree, err := ast(nodeGroups[i].tokens)
		if err != nil {
			return nil, err
		}
		nodeGroups[i].tree = tree
		tokens, err = tree.parse(mapParams)
		if err != nil {
			return nil, err
		}
		nodeGroups[i].tokens = tokens
		generatedTokens = append(generatedTokens, nodeGroups[i].tokens...)
	}

	return generatedTokens, nil
}

func recursiveIfnestt(tokens []token, idx *int, mapParams map[string]interface{}) ([]token, error) {
	tmp := []token{}
	// ifstack := []tokenKind{}
	iftokenGroup := []token{tokens[*idx]} // IF
	*idx++
	for {
		if tokens[*idx].kind == tkIf {
			nestTokens, err := recursiveIfnestt(tokens, idx, mapParams)
			if err != nil {
				return nil, err
			}
			tmp = append(tmp, nestTokens...)
			// *idx++
			continue
		}
		if tokens[*idx].kind == tkElse || tokens[*idx].kind == tkElif {
			nt, err := splitNodeGroup(tmp, mapParams)
			if err != nil {
				return nil, err
			}
			tmp = nt
			tmp = append(tmp, tokens[*idx])
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

		iftokenGroup = append(iftokenGroup, tmp...)                      // IF
		iftokenGroup = append(iftokenGroup, tokens[*idx])                // END
		iftokenGroup = append(iftokenGroup, token{kind: tkEndOfProgram}) // 分割ごとに追加
		tmp = []token{}
		*idx++
		break
	}

	tree, err := ast(iftokenGroup)
	if err != nil {
		return nil, err
	}
	generatedTokens, err := tree.parse(mapParams)
	if err != nil {
		return nil, err
	}

	return generatedTokens, nil
}

func (n *nodeGroup) parse(params map[string]interface{}) ([]token, error) {
	return genInner(n.tree, params)
}

func (n *nodeGroup) String() string {
	return fmt.Sprintf("typ: %v, tokens: %v", n.typ, n.tokens)
}
