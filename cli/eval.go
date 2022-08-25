package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/fatih/color"
	"github.com/goccy/go-yaml"

	"github.com/future-architect/go-twowaysql"
)

func eval(srcPath string, params []string) error {
	stat, _ := os.Stdin.Stat()
	var finalParams map[string]any
	var err error
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		finalParams, err = parseParams(params, os.Stdout)
	} else {
		finalParams, err = parseParams(params, nil)
	}
	if err != nil {
		return err
	}

	srcSql, err := readSql(srcPath)
	if err != nil {
		return err
	}

	convertedSrc, sqlParams, err := twowaysql.Eval(srcSql, finalParams)
	if err != nil {
		return err
	}
	title := color.New(color.FgHiRed, color.Bold)
	title.Println("# Converted Source")
	fmt.Printf("\n")
	quick.Highlight(os.Stdout, convertedSrc, "sql", "terminal", "monokai")
	title.Println("\n# Parameters")
	fmt.Printf("\n")
	sqlParamYaml, _ := yaml.Marshal(sqlParams)
	quick.Highlight(os.Stdout, string(sqlParamYaml), "yaml", "terminal", "monokai")
	return nil
}
