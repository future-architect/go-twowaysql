package cli

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/future-architect/go-twowaysql"
	"github.com/jmoiron/sqlx"
	"github.com/shibukawa/formatdata-go"
	"golang.org/x/crypto/ssh/terminal"
)

type explainConfig struct {
	Statement        string
	ResultIsTable    bool
	RollbackRequired bool
}

var explainStatements = map[string]*explainConfig{
	"pgx":      {"EXPLAIN ANALYZE ", false, true},
	"postgres": {"EXPLAIN ANALYZE ", false, true},
	"sqlite":   {"EXPLAIN QUERY PLAN ", true, false},
	"sqlite3":  {"EXPLAIN QUERY PLAN ", true, false},
	"mysql":    {"EXPLAIN ", true, false},
}

var outputFormats = map[string]formatdata.OutputFormat{
	"default": formatdata.Terminal,
	"md":      formatdata.Markdown,
	"json":    formatdata.JSON,
	"yaml":    formatdata.YAML,
}

func run(driver, dbSrc, srcFilePath string, params []string, explain, rollback bool, outputFormat string, out io.Writer) error {
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

	srcSql, err := readSql(srcFilePath)
	if err != nil {
		return err
	}

	var result []map[string]any

	// todo: time limit param
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var explainStatement *explainConfig

	if explain {
		var ok bool
		if explainStatement, ok = explainStatements[driver]; ok {
			rollback = explainStatement.RollbackRequired
			srcSql = explainStatement.Statement + srcSql
		} else {
			return fmt.Errorf("unknown driver to execute explain. pull request is welcome to add support to twowaysql: %s", driver)
		}
	}

	db, err := sqlx.Open(driver, dbSrc)
	if err != nil {
		return err
	}
	tws := twowaysql.New(db)
	defer tws.Close()

	start := time.Now()
	if rollback {
		tr, err := tws.Begin(ctx)
		if err != nil {
			return err
		}
		defer tr.Rollback()
		if useQuery(srcSql, explain) {
			err = tr.Select(ctx, &result, srcSql, finalParams)
		} else {
			result, err = mapResult(tr.Exec(ctx, srcSql, finalParams))
		}
		if err != nil {
			return err
		}
	} else {
		if useQuery(srcSql, explain) {
			err = tws.Select(ctx, &result, srcSql, finalParams)
		} else {
			result, err = mapResult(tws.Exec(ctx, srcSql, finalParams))
		}
		if err != nil {
			return err
		}
	}

	duration := time.Now().Sub(start)

	if explainStatement != nil && !explainStatement.ResultIsTable {
		dumpResult(result, out)
	} else {
		if len(result) > 0 {
			if out != nil {
				formatdata.FormatDataTo(result, out, formatdata.Opt{
					OutputFormat: outputFormats[outputFormat],
				})
			} else {
				formatdata.FormatData(result, formatdata.Opt{
					OutputFormat: outputFormats[outputFormat],
				})
			}
		}
		if !explain && isTerminal(out) {
			color.HiRed("\nQuery takes %v\n", duration)
		}
	}

	return nil
}

func dumpResult(result []map[string]any, out io.Writer) {
	var builder strings.Builder
	var key string
	for _, row := range result {
		for k, v := range row {
			key = k
			if line, ok := v.(string); ok {
				builder.WriteString(line)
				builder.WriteByte('\n')
			}
		}
	}
	if isTerminal(out) {
		title := color.New(color.FgHiRed, color.Bold)
		title.Printf("# %s\n\n", key)
		color.Yellow(builder.String())
	} else {
		fmt.Println(builder.String())
	}
}

func isTerminal(out io.Writer) bool {
	if out == nil {
		return true
	}
	o, ok := out.(*os.File)
	return ok && terminal.IsTerminal(int(o.Fd()))
}

var splitter = regexp.MustCompile(`\s+`)

func useQuery(sql string, explain bool) bool {
	if explain {
		return true
	}
	for _, w := range splitter.Split(sql, -1) {
		if w == "" {
			continue
		}
		if strings.ToLower(w) == "select" {
			return true
		} else {
			return false
		}
	}
	return false
}

func mapResult(dbResult sql.Result, err error) ([]map[string]any, error) {
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	lastInsertId, err := dbResult.LastInsertId()
	// not all driver support these values
	if err == nil {
		result["Last Insert Id"] = lastInsertId
	}
	rowsAffected, err := dbResult.RowsAffected()
	if err == nil {
		result["Rows Affected"] = rowsAffected
	}
	if len(result) > 0 {
		return []map[string]any{
			result,
		}, nil
	}
	return nil, nil
}
