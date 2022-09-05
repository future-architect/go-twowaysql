package twowaysql

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/shibukawa/mdd-go"
)

// ParamType is for describing parameters
type ParamType int

const (
	InvalidType ParamType = iota
	BoolType
	ByteType
	FloatType
	IntType
	TextType
	TimestampType
)

var paramTypeMap = map[string]ParamType{
	"text":      TextType,
	"string":    TextType,
	"str":       TextType,
	"varchar":   TextType,
	"integer":   IntType,
	"int":       IntType,
	"float":     FloatType,
	"float64":   FloatType,
	"double":    FloatType,
	"bool":      BoolType,
	"boolean":   BoolType,
	"time":      TimestampType,
	"timestamp": TimestampType,
	"byte":      ByteType,
	"tinyint":   ByteType,
}

func (p ParamType) String() string {
	switch p {
	case InvalidType:
		return "invalid"
	case BoolType:
		return "bool"
	case ByteType:
		return "byte"
	case FloatType:
		return "float"
	case IntType:
		return "integer"
	case TextType:
		return "text"
	case TimestampType:
		return "timestamp"
	default:
		return "unknown"
	}
}

func (p ParamType) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *ParamType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("data should be a string, got %s", data)
	}
	pt, ok := paramTypeMap[s]
	if !ok {
		return fmt.Errorf("invalid UserRole %s", s)
	}
	*p = pt
	return nil
}

// Document contains SQL and metadata
type Document struct {
	SQL        string       `json:"sql"`
	Title      string       `json:"title"`
	Params     []Param      `json:"params"`
	CRUDMatrix []CRUDMatrix `json:"crud_matrix,omitempty"`
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

var docJig = mdd.NewDocJig[Document]()

func init() {
	docJig.Alias("Title").Lang("ja", "タイトル")

	docJig.Alias("Parameters", "Parameter").Lang("ja", "パラメータ", "引数")
	docJig.Alias("CRUD Matrix", "CRUD Table").Lang("ja", "CRUDマトリックス", "CRUD図", "CRUD表")

	docJig.Alias("Name").Lang("ja", "パラメータ名", "名前")
	docJig.Alias("Type").Lang("ja", "型", "タイプ")
	docJig.Alias("Description", "Desc", "Detail").Lang("ja", "説明", "詳細")

	docJig.Alias("Table").Lang("ja", "テーブル")

	root := docJig.Root().Label("Title")

	root.CodeFence("SQL", "sql").SampleCode("select * from table where id = 1;")

	params := root.Child(".", "Parameters").Table("Params")
	params.Field("Name", "Name").Required()
	params.Field("Type").Required().As(func(typeName string, d *Document) (any, error) {
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
}

// ParseMarkdownFile parses markdown file
func ParseMarkdownFile(filepath string) (*Document, error) {
	return docJig.ParseFile(filepath)
}

// ParseMarkdown parses markdown content
func ParseMarkdown(r io.Reader) (*Document, error) {
	return docJig.Parse(r)
}

// ParseMarkdown parses markdown content
func ParseMarkdownString(src string) (*Document, error) {
	return docJig.ParseString(src)
}

func GenerateMarkdown(w io.Writer, lang string) error {
	return docJig.GenerateTemplate(w, mdd.GenerateOption{
		Language: lang,
	})
}
