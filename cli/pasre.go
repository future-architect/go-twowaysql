package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/future-architect/go-twowaysql"
	"gopkg.in/yaml.v2"
)

func parseFile(srcPath, dumpFormat string) error {
	switch filepath.Ext(srcPath) {
	case ".md":
		return parseMarkdownFile(srcPath, dumpFormat)
	case ".sql":
		return parseSQLFile(srcPath, dumpFormat)
	default:
		return fmt.Errorf("parse command only supports .sql/.md file, but %s", filepath.Ext(srcPath))
	}
}

type sqlParams struct {
	Params []string `json:"params"`
}

func parseSQLFile(srcPath, dumpFormat string) error {
	src, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	// need more clever function...
	params := make(map[string]any)
	var paramNames []string
	prefix := "no parameter that matches the bind value: "
	for {
		_, _, err := twowaysql.Eval(string(src), params)
		if err == nil {
			break
		}
		if strings.HasPrefix(err.Error(), prefix) {
			param := strings.TrimPrefix(err.Error(), prefix)
			paramNames = append(paramNames, param)
			params[param] = ""
		} else {
			return fmt.Errorf("parse error: %w", err)
		}
	}
	return dump(&sqlParams{Params: paramNames}, dumpFormat)
}

func parseMarkdownFile(srcPath, dumpFormat string) error {
	doc, err := twowaysql.ParseMarkdownFile(srcPath)
	if err != nil {
		return err
	}
	return dump(doc, dumpFormat)
}

func dump(src any, dumpFormat string) error {
	if isTerminal(os.Stdout) {
		if dumpFormat == "json" {
			var b bytes.Buffer
			e := json.NewEncoder(&b)
			e.SetIndent("", "  ")
			e.Encode(src)
			return quick.Highlight(os.Stdout, b.String(), "json", "terminal16", "github")
		} else {
			var b bytes.Buffer
			e := yaml.NewEncoder(&b)
			e.Encode(src)
			return quick.Highlight(os.Stdout, b.String(), "yaml", "terminal16", "github")
		}
	} else {
		if dumpFormat == "json" {
			e := json.NewEncoder(os.Stdout)
			e.SetIndent("", "  ")
			return e.Encode(src)
		} else {
			e := yaml.NewEncoder(os.Stdout)
			return e.Encode(src)
		}
	}
}
