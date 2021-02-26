package twowaysql

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

type Person struct {
	FirstName string `twowaysql: "first_name"`
	LastName  string `twowaysql: "last_name"`
	Email     string `twowaysql: "email"`
}

// inputStructは本来structだが現時点ではpeople決め打ちで書く
type Twowaysql struct {
	db             *sql.DB
	query          string
	convertedQuery string
	params         map[string]interface{}
	inputStructs   *[]Person
}

func New(db *sql.DB) *Twowaysql {
	return &Twowaysql{
		db: db,
	}
}

// ConvertedQuery returns convertedQuery property
func (t *Twowaysql) ConvertedQuery() string {
	return t.convertedQuery
}

func (t *Twowaysql) withQuery(query string) *Twowaysql {
	return &Twowaysql{
		db:             t.db,
		query:          query,
		convertedQuery: t.convertedQuery,
		params:         t.params,
		inputStructs:   t.inputStructs,
	}
}

func (t *Twowaysql) withParams(params map[string]interface{}) *Twowaysql {
	return &Twowaysql{
		db:             t.db,
		query:          t.query,
		convertedQuery: t.convertedQuery,
		params:         params,
		inputStructs:   t.inputStructs,
	}
}

func (t *Twowaysql) withInputStruct(inputStructs *[]Person) *Twowaysql {
	return &Twowaysql{
		db:             t.db,
		query:          t.query,
		convertedQuery: t.convertedQuery,
		params:         t.params,
		inputStructs:   inputStructs,
	}
}

// 事前条件: inputStructのフィールドとqueryで返ってくる要素の長さと並びは一致していなければならない。
func (t *Twowaysql) Select(inputStructs *[]Person, query string, params map[string]interface{}) *Twowaysql {
	return t.withParams(params).withQuery(query).withInputStruct(inputStructs)
}

func (t *Twowaysql) Run(ctx context.Context) error {
	convertedQuery, err := t.parse()
	if err != nil {
		return err
	}
	t.convertedQuery = convertedQuery

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
	rows, err := t.db.QueryContext(ctx, convertedQuery, params...)
	if err != nil {
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

	return rows.Err()
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
	for i := 0; i < count; i++ {
		str = strings.Replace(str, "?", "$"+strconv.Itoa(i+1), 1)
	}
	return str
}
