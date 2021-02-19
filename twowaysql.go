package twowaysql

type twowaysql struct {
	params map[string]interface{}
}

func New() twowaysql {
	return twowaysql{
		params: map[string]interface{}{},
	}
}

func (t twowaysql) WithParams(params map[string]interface{}) twowaysql {
	return twowaysql{
		params: params,
	}
}

//内部でバインドパラメータのデータを持っている必要があるかも
func (t twowaysql) convert(query string) (string, error) {
	tokens, err := tokinize(query)
	if err != nil {
		return "", err
	}
	tree, err := ast(tokens)
	if err != nil {
		return "", err
	}
	convertedStr, err := gen(tree)
	if err != nil {
		return "", err
	}

	return convertedStr, nil
}
