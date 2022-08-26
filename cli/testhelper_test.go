package cli

import (
	"regexp"
	"strings"
)

var stripSpacePattern = regexp.MustCompile("(^[ \t]*)")

func trimIndent(src string) string {
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
