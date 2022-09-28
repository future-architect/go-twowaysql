package twowaysql

import (
	"encoding/json"
	"fmt"
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

type MatchRule int

const (
	SelectExactMatch MatchRule = iota + 1
	SelectMatch
	ExecExactMatch
	ExecMatch
)

var matchRuleMap = map[string]MatchRule{
	"select(exact-order)": SelectExactMatch,
	"select":              SelectExactMatch,
	"select(free-order)":  SelectMatch,
	"exec(exact-order)":   ExecExactMatch,
	"exec(free-order)":    ExecMatch,
	"exec":                ExecMatch,
}

func (m MatchRule) String() string {
	switch m {
	case SelectExactMatch:
		return "select(exact-order)"
	case SelectMatch:
		return "select(free-order)"
	case ExecExactMatch:
		return "exec(exact-order)"
	case ExecMatch:
		return "exec(free-order)"
	default:
		return ""
	}
}

func (m MatchRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *MatchRule) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("data should be a string, got %s", data)
	}
	mr, ok := matchRuleMap[s]
	if !ok {
		return fmt.Errorf("invalid UserRole %s", s)
	}
	*m = mr
	return nil
}
