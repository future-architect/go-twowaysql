package testhelper

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

var stripSpacePattern = regexp.MustCompile("(^[ \t]*)")

func TrimIndent(t *testing.T, src string) string {
	t.Helper()
	lines := strings.Split(src, "\n")
	if lines[0] == "" {
		lines = lines[1:]
	}
	matches := stripSpacePattern.FindStringSubmatch(lines[0])
	var b strings.Builder
	for i, line := range lines {
		b.WriteString(strings.TrimPrefix(line, matches[0]))
		if i != len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func SourceStr(t *testing.T) string {
	t.Helper()
	if host, ok := os.LookupEnv("POSTGRES_HOST"); ok {
		return fmt.Sprintf("host=%s user=postgres password=postgres dbname=postgres sslmode=disable", host)
	} else {
		return "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"
	}
}
