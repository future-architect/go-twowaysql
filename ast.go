package twowaysql

type NodeKind int

const (
	NdSQLstmt NodeKind = iota + 1
	NdIf
	NdElif
	NdElse
	NdEnd
)

type Tree struct {
	Kind  NodeKind
	Left  *Tree
	Right *Tree
	Token *Token
}

func ast(tokens []Token) (*Tree, error) {
	return &Tree{}, nil
}
