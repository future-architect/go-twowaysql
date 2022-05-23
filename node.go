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

func splitNodeGroup(tokens []token) ([]nodeGroup, error) {
	var res []nodeGroup
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
		res = append(res, ng)

		tmp = []token{}
		for {
			if tokens[idx].kind != tkEnd {
				tmp = append(tmp, tokens[idx])
				idx++
				continue
			}
			tmp = append(tmp, tokens[idx])
			tmp = append(tmp, token{kind: tkEndOfProgram}) // 分割ごとに追加
			ng = nodeGroup{
				tokens: tmp,
			}
			res = append(res, ng)
			tmp = []token{}
			idx++
			break
		}
	}
	if len(tokens) != 0 {
		if tokens[len(tokens)-1].kind != tkEndOfProgram {
			tmp = append(tmp, token{kind: tkEndOfProgram}) // 分割ごとに追加
		}
		res = append(res, nodeGroup{tokens: tmp})
	}

	for _, v := range res {
		fmt.Printf("ngs: %v\n", v)
	}

	for i := range res {
		tree, err := ast(res[i].tokens)
		if err != nil {
			return nil, err
		}
		res[i].tree = tree
	}

	return res, nil
}

func (n *nodeGroup) parse(params map[string]interface{}) ([]token, error) {
	return genInner(n.tree, params)
}

func (n *nodeGroup) String() string {
	return fmt.Sprintf("typ: %v, tokens: %v", n.typ, n.tokens)
}
