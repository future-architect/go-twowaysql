package cli

import (
	"os"
	"strings"

	"github.com/future-architect/go-twowaysql"
)

func readSql(srcPath string) (sql string, err error) {
	if strings.HasSuffix(srcPath, ".md") {
		doc, err := twowaysql.ParseMarkdownFile(srcPath)
		if err != nil {
			return "", err
		}
		return doc.SQL, err
	}
	src, err := os.ReadFile(srcPath)
	if err != nil {
		return "", err
	}

	return string(src), nil
}
