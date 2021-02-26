package twowaysql

// TODO: parseしているのに戻り値がstringなのが違和感、良い命名を考える。
func (t *Twowaysql) parse() (string, error) {
	tokens, err := tokinize(t.query)
	if err != nil {
		return "", err
	}
	tree, err := ast(tokens)
	if err != nil {
		return "", err
	}

	return gen(tree, t.params)
}
