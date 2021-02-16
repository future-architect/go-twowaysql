package twowaysql

type TokenKind int

const (
	TkSQLStmt TokenKind = iota
	TkIf
	TkElif
	TkElse
	TkBind
)

type Token struct {
	kind TokenKind
	str  string
}

func tokinize(str string) []Token {
	return []Token{}
}
