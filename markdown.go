package twowaysql

import (
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"

	"github.com/shibukawa/mdd-go"
	"gopkg.in/yaml.v2"
)

// Document contains SQL and metadata
type Document struct {
	SQL               string       `json:"sql"`
	Title             string       `json:"title"`
	Params            []Param      `json:"params"`
	CRUDMatrix        []CRUDMatrix `json:"crud_matrix,omitempty"`
	TestCases         []TestCase   `json:"testcases,omitempty"`
	CommonTestFixture Fixture      `json:"common_test_fixtures,omitempty"`
}

type Fixture struct {
	Lang   string
	Code   string
	Tables []Table
}

// document absorb diffs between raw Markdown representation and public Document type
type document struct {
	SQL                      string
	Title                    string
	Params                   []Param
	CRUDMatrix               []CRUDMatrix
	TestCases                []testCase
	RawCommonTestFixture     string
	RawCommonTestFixtureLang string
	parsedCommonTestFixture  []Table
}

// Param is parameter type of 2-Way-SQL
type Param struct {
	Name        string    `json:"name"`
	Type        ParamType `json:"type"`
	Value       string    `json:"value"`
	Description string    `json:"description,omitempty"`
}

// CRUDMatrix represents CRUD Matrix
type CRUDMatrix struct {
	Table       string `json:"table"`
	C           bool   `json:"c"`
	R           bool   `json:"r"`
	U           bool   `json:"u"`
	D           bool   `json:"d"`
	Description string `json:"description,omitempty"`
}

type Table struct {
	MatchRule MatchRule  `json:"type"`
	Name      string     `json:"name"`
	Cells     [][]string `json:"cells"`
}

type TestCase struct {
	Name      string
	Params    map[string]string
	TestQuery string
	Expect    [][]string
	Fixtures  []Table
}

type testCase struct {
	Name            string
	RawTest         string
	parsedFixtures  []Table
	parsedTestQuery string
	parsedExpect    [][]string
	parsedParams    map[string]string
}

func parseFixture(src string) (map[string][][]string, bool) {
	tempSliceYaml := struct {
		Tables map[string][][]string `yaml:"fixtures"`
	}{}
	tempMapYaml := struct {
		Tables map[string][]map[string]string `yaml:"fixtures"`
	}{}
	if err := yaml.Unmarshal([]byte(src), &tempSliceYaml); err == nil {
		return tempSliceYaml.Tables, true
	} else if err := yaml.Unmarshal([]byte(src), &tempMapYaml); err == nil {
		result := make(map[string][][]string)
		for k, v := range tempMapYaml.Tables {
			result[k] = convertTableMapToSlice(v)
		}
		return result, true
	}
	return nil, false
}

func parseExpect(src string) ([][]string, string, map[string]string, bool) {
	tempSliceYaml := struct {
		Param     map[string]string `yaml:"params"`
		TestQuery string            `yaml:"testQuery"`
		Expect    [][]string        `yaml:"expect"`
	}{}
	tempMapYaml := struct {
		Param     map[string]string   `yaml:"params"`
		TestQuery string              `yaml:"testQuery"`
		Expect    []map[string]string `yaml:"expect"`
	}{}
	if err := yaml.Unmarshal([]byte(src), &tempSliceYaml); err == nil {
		return tempSliceYaml.Expect, tempSliceYaml.TestQuery, tempSliceYaml.Param, true
	} else if err := yaml.Unmarshal([]byte(src), &tempMapYaml); err == nil {
		return convertTableMapToSlice(tempMapYaml.Expect), tempMapYaml.TestQuery, tempMapYaml.Param, true
	}
	return nil, "", nil, false
}

var (
	acceptableKeysInGlobalFixture = map[string]bool{
		"fixtures": true,
	}
	acceptableKeysInLocalTestCases = map[string]bool{
		"fixtures":  true,
		"params":    true,
		"testQuery": true,
		"expect":    true,
	}
)

func checkKeys(src string, acceptableKeys map[string]bool, label string) error {
	var temp map[string]any
	err := yaml.Unmarshal([]byte(src), &temp)
	if err != nil {
		return err
	}
	var mismatch []string
	for key := range temp {
		if _, ok := acceptableKeys[key]; !ok {
			mismatch = append(mismatch, key)
		}
	}
	if len(mismatch) > 0 {
		var keys []string
		for key := range acceptableKeys {
			keys = append(keys, key)
		}
		sort.Strings(mismatch)
		sort.Strings(keys)
		return fmt.Errorf("YAML keys %s is invalid in %s (%s are acceptable)", strings.Join(mismatch, ", "), label, strings.Join(keys, ", "))
	}
	return nil
}

func (d *document) PostProcess() error {
	if d.RawCommonTestFixtureLang == "yaml" {
		err := checkKeys(d.RawCommonTestFixture, acceptableKeysInGlobalFixture, d.Title)
		if err != nil {
			return err
		}
		if parsed, ok := parseFixture(d.RawCommonTestFixture); ok {
			for k, cells := range parsed {
				d.parsedCommonTestFixture = append(d.parsedCommonTestFixture, Table{
					Name:  k,
					Cells: cells,
				})
			}
		}
	}
	for i, tc := range d.TestCases {
		err := checkKeys(tc.RawTest, acceptableKeysInLocalTestCases, tc.Name+" of "+d.Title)
		if err != nil {
			return err
		}
		if parsed, ok := parseFixture(tc.RawTest); ok {
			for k, cells := range parsed {
				tc.parsedFixtures = append(tc.parsedFixtures, Table{
					Name:  k,
					Cells: cells,
				})
			}
		}
		if parsed, testQuery, params, ok := parseExpect(tc.RawTest); ok {
			tc.parsedExpect = parsed
			tc.parsedTestQuery = testQuery
			tc.parsedParams = params
		} else {
			return fmt.Errorf("can't parse yaml of test '%s'", tc.Name)
		}
		d.TestCases[i] = tc
	}
	return nil
}

func (d document) ToDocument() *Document {
	result := &Document{
		SQL:        d.SQL,
		Title:      d.Title,
		Params:     d.Params,
		CRUDMatrix: d.CRUDMatrix,
	}
	switch d.RawCommonTestFixtureLang {
	case "yaml":
		result.CommonTestFixture = Fixture{
			Lang:   "yaml",
			Tables: d.parsedCommonTestFixture,
		}
	case "sql":
		result.CommonTestFixture = Fixture{
			Lang: "sql",
			Code: d.RawCommonTestFixture,
		}
	}
	for _, tc := range d.TestCases {
		result.TestCases = append(result.TestCases, TestCase{
			Name:      tc.Name,
			Params:    tc.parsedParams,
			TestQuery: tc.parsedTestQuery,
			Expect:    tc.parsedExpect,
			Fixtures:  tc.parsedFixtures,
		})
	}

	return result
}

var docJig = mdd.NewDocJig[document]()

func init() {
	docJig.Alias("Title").Lang("ja", "タイトル")

	docJig.Alias("Parameters", "Parameter").Lang("ja", "パラメータ", "引数")
	docJig.Alias("CRUD Matrix", "CRUD Table").Lang("ja", "CRUDマトリックス", "CRUD図", "CRUD表")

	docJig.Alias("Name").Lang("ja", "パラメータ名", "名前")
	docJig.Alias("Type").Lang("ja", "型", "タイプ")
	docJig.Alias("Description", "Desc", "Detail").Lang("ja", "説明", "詳細")

	docJig.Alias("Table").Lang("ja", "テーブル")

	docJig.Alias("Test", "Tests", "Sample", "Samples", "Example", "Examples").Lang("ja", "テスト", "サンプル", "実行例")
	docJig.Alias("Case", "Test Case", "TestCase").Lang("ja", "ケース", "テストケース")

	root := docJig.Root().Label("Title")

	root.CodeFence("SQL", "sql").SampleCode("select * from table where id = 1;")

	params := root.Child(".", "Parameters").Table("Params")
	params.Field("Name", "Name").Required()
	params.Field("Type").Required().As(func(typeName string, d *document) (any, error) {
		t, ok := paramTypeMap[strings.ToLower(typeName)]
		if ok {
			return t, nil
		}
		return nil, fmt.Errorf("type '%s' is invalid", typeName)
	})
	params.Field("Description")

	crudMatrix := root.Child(".", "CRUD Matrix").Table("CRUDMatrix")
	crudMatrix.Field("Table").Required()
	crudMatrix.Field("C").Required()
	crudMatrix.Field("R").Required()
	crudMatrix.Field("U").Required()
	crudMatrix.Field("D").Required()
	crudMatrix.Field("Description")

	test := root.Child(".", "Test")
	test.CodeFence("RawCommonTestFixture", "sql", "yaml").Language("RawCommonTestFixtureLang")
	testcase := test.Children("TestCases", "Case")
	testcase.Label("Name")
	testcase.CodeFence("RawTest", "yaml")
}

// ParseMarkdownFile parses markdown file
func ParseMarkdownFile(filepath string) (*Document, error) {
	d, err := docJig.ParseFile(filepath)
	if err != nil {
		return nil, err
	}
	return d.ToDocument(), err
}

// ParseMarkdown parses markdown content
func ParseMarkdown(r io.Reader) (*Document, error) {
	d, err := docJig.Parse(r)
	if err != nil {
		return nil, err
	}
	return d.ToDocument(), err
}

// ParseMarkdown parses markdown content
func ParseMarkdownString(src string) (*Document, error) {
	d, err := docJig.ParseString(src)
	if err != nil {
		return nil, err
	}
	return d.ToDocument(), err
}

// ParseMarkdownGlob parses markdown files to match patterns
func ParseMarkdownGlob(pattern ...string) (map[string]*Document, error) {
	ds, err := docJig.ParseGlob(pattern...)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Document)
	for k, d := range ds {
		result[k] = d.ToDocument()
	}
	return result, err
}

// ParseMarkdown parses markdown content
func ParseMarkdownFS(fsys fs.FS, pattern ...string) (map[string]*Document, error) {
	ds, err := docJig.ParseFS(fsys, pattern...)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*Document)
	for k, d := range ds {
		result[k] = d.ToDocument()
	}
	return result, err
}

// ParseMarkdown parses markdown content
func GenerateMarkdown(w io.Writer, lang string) error {
	return docJig.GenerateTemplate(w, mdd.GenerateOption{
		Language: lang,
	})
}

func convertTableMapToSlice(table []map[string]string) [][]string {
	var headers []string
	existingCheck := map[string]bool{}
	for _, row := range table {
		for k := range row {
			if !existingCheck[k] {
				headers = append(headers, k)
				existingCheck[k] = true
			}
		}
	}
	sort.Strings(headers)

	slices := make([][]string, len(table)+1)
	slices[0] = headers
	for r, row := range table {
		rowSlice := make([]string, len(headers))
		for c, h := range headers {
			if v, ok := row[h]; ok {
				rowSlice[c] = v
			} else {
				rowSlice[c] = ""
			}
		}
		slices[r+1] = rowSlice
	}
	return slices
}
