package twowaysql

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
)

type Person struct {
	FirstName string `twowaysql: "first_name"`
	LastName  string `twowaysql: "last_name"`
	Email     string `twowaysql: "email"`
}

// inputStructは本来structだが現時点ではpeople決め打ちで書く
type twowaysql struct {
	query        string
	params       map[string]interface{}
	inputStructs *[]Person
}

func new() *twowaysql {
	return &twowaysql{}
}

func (t *twowaysql) withQuery(query string) *twowaysql {
	return &twowaysql{
		query:        query,
		params:       t.params,
		inputStructs: t.inputStructs,
	}
}

func (t *twowaysql) withParams(params map[string]interface{}) *twowaysql {
	return &twowaysql{
		query:        t.query,
		params:       params,
		inputStructs: t.inputStructs,
	}
}

func (t *twowaysql) withInputStruct(inputStructs *[]Person) *twowaysql {
	return &twowaysql{
		query:        t.query,
		params:       t.params,
		inputStructs: inputStructs,
	}
}

func (t *twowaysql) Run(db *sql.DB, ctx context.Context) error {
	convertedQuery, err := t.convert()
	if err != nil {
		return err
	}

	binds, err := retrieveBinds(convertedQuery)
	if err != nil {
		return err
	}

	var params []interface{}
	for _, bind := range binds {
		if elem, ok := t.params[bind]; ok {
			params = append(params, elem)
		} else {
			return errors.New("no parameter that matches the bind value")
		}
	}

	//一時的な措置、本当はどこかでdatabaseのtypeを知る必要がある。
	postgres := true
	if postgres {
		convertedQuery = convertPlaceHolder(convertedQuery)
	}

	//log.Println("query", convertedQuery)
	//log.Println(params...)
	rows, err := db.QueryContext(ctx, convertedQuery, params...)
	if err != nil {
		log.Println("ERROR")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		p := Person{}
		if err := rows.Scan(&p.FirstName, &p.LastName, &p.Email); err != nil {
			return err
		}
		p.FirstName = strings.TrimRight(p.FirstName, " ")
		p.LastName = strings.TrimRight(p.LastName, " ")
		p.Email = strings.TrimRight(p.Email, " ")
		*t.inputStructs = append(*t.inputStructs, p)
	}

	rerr := rows.Close()
	if rerr != nil {
		return rerr
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func (t *twowaysql) convert() (string, error) {
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

// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func Select(inputStructs *[]Person, query string, params map[string]interface{}) *twowaysql {
	t := new().withParams(params).withQuery(query).withInputStruct(inputStructs)
	return t
}

// ?/*...*/を抽出する
func retrieveBinds(query string) ([]string, error) {
	var binds []string
	tokens, err := tokinize(query)
	if err != nil {
		return nil, err
	}
	for _, token := range tokens {
		if token.kind == TkBind {
			// ?/*...*/という形を想定
			bindValue := retrieveBind(token.str)
			binds = append(binds, bindValue)
		}
	}
	return binds, nil
}

// ?/* value */ からvalueを取り出す
// 事前条件: strは?/* ... */という形である
func retrieveBind(str string) string {
	var retStr string
	retStr = strings.Trim(str, " ")
	retStr = strings.TrimLeft(retStr, "?")
	retStr = strings.TrimPrefix(retStr, "/*")
	retStr = strings.TrimSuffix(retStr, "*/")
	return strings.Trim(retStr, " ")
}

// ?/* ... */ を $1/* ... */のような形に変換する。
func convertPlaceHolder(str string) string {
	count := strings.Count(str, "?")
	log.Println("count", count)
	for i := 0; i < count; i++ {
		str = strings.Replace(str, "?", "$"+strconv.Itoa(i+1), 1)
	}
	return str
}
