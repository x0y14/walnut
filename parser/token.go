package parser

type TokenKind int

const (
	Illegal TokenKind = iota
	Eof
	Whitespace

	Ident   // events
	String  // "str"
	Integer // 123
	Float   // 12.3

	Lcb // {
	Rcb // }

	Assign // =
)

type Token struct {
	Pos  *Position
	Kind TokenKind
	Raw  []rune
	Next *Token
}

func NewToken(pos *Position, kind TokenKind, raw []rune) *Token {
	return &Token{
		Pos:  pos,
		Kind: kind,
		Raw:  raw,
		Next: nil,
	}
}
