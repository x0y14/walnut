package parser

import "fmt"

type LexOptions struct {
	Clean bool
}

type Lexer struct {
	chars []rune
	pos   int
	curt  *Token
}

func (l *Lexer) isEof() bool {
	return len(l.chars) <= l.pos
}

func (l *Lexer) charIs(r ...rune) bool {
	for _, r_ := range r {
		if r_ == l.chars[l.pos] {
			return true
		}
	}
	return false
}
func (l *Lexer) charIsAlpha() bool {
	if 'a' <= l.chars[l.pos] && l.chars[l.pos] <= 'z' {
		return true
	}
	if 'A' <= l.chars[l.pos] && l.chars[l.pos] <= 'Z' {
		return true
	}
	return false
}

func (l *Lexer) consume(r ...rune) (rune, bool) {
	if len(r) == 0 {
		_r := l.chars[l.pos]
		l.pos++
		return _r, true
	}
	if l.charIs(r...) {
		_r := l.chars[l.pos]
		l.pos++
		return _r, true
	}
	return 0, false
}

func (l *Lexer) expect(r rune) error {
	if l.charIs(r) {
		l.pos++
		return nil
	}
	return fmt.Errorf("unexpected char: want=%v, got=%v", r, l.chars[l.pos])
}

// Lexerにハンドラを登録
func (l *Lexer) getHandler() TokenHandler {
	switch {
	case l.charIs(' ', '\n', '\r', '\t'):
		return &WhitespaceHandler{}
	case l.charIs('"'):
		return &StringHandler{}
	case l.charIs('0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-'):
		return &NumericHandler{}
	case l.charIs('{', '}', '='):
		return &SymbolHandler{}
	case l.charIsAlpha() || l.charIs('_'):
		return &IdentHandler{}
	default:
		return nil
	}
}

func (l *Lexer) Lex(opts *LexOptions, script string) (*Token, error) {
	l.chars = []rune(script)
	l.pos = 0
	head := NewToken(nil, Illegal, nil)
	l.curt = head

	for !l.isEof() {
		handler := l.getHandler()
		if handler == nil {
			return nil, fmt.Errorf("unexpected char: at=%v", l.pos)
		}
		if err := handler.Handle(l, opts); err != nil {
			return nil, err
		}
	}

	eof := NewToken(NewPosition(l.pos, l.pos), Eof, nil)
	l.curt.Next = eof
	l.curt = l.curt.Next
	return head.Next, nil
}

type TokenHandler interface {
	Handle(l *Lexer, opts *LexOptions) error
}

type WhitespaceHandler struct{}

func (h *WhitespaceHandler) Handle(l *Lexer, opts *LexOptions) error {
	pos := NewPosition(l.pos, 0)
	var ws []rune

	for !l.isEof() {
		r, ok := l.consume(' ', '\n', '\r', '\t')
		if ok {
			ws = append(ws, r)
		} else {
			break
		}
	}

	pos.EndedAt = l.pos

	if !opts.Clean {
		l.curt.Next = NewToken(pos, Whitespace, ws)
		l.curt = l.curt.Next
	}
	return nil
}

type IdentHandler struct{}

func (i *IdentHandler) Handle(l *Lexer, opts *LexOptions) error {
	_ = opts
	pos := NewPosition(l.pos, 0)
	var ident []rune

	for !l.isEof() {
		if l.charIsAlpha() || l.charIs('_') {
			r, _ := l.consume()
			ident = append(ident, r)
		} else {
			break
		}
	}

	pos.EndedAt = l.pos
	l.curt.Next = NewToken(pos, Ident, ident)
	l.curt = l.curt.Next
	return nil
}

type StringHandler struct{}

func (s StringHandler) Handle(l *Lexer, opts *LexOptions) error {
	_ = opts
	pos := NewPosition(l.pos, 0)
	var str []rune

	// opening "
	if err := l.expect('"'); err != nil {
		return err
	}

strLoop:
	for !l.isEof() {
		// closing "
		if l.charIs('"') {
			break strLoop
		}
		r, _ := l.consume()
		str = append(str, r)
	}
	if l.isEof() {
		return fmt.Errorf("string without closing dq: at=%v", l.pos)
	}
	if err := l.expect('"'); err != nil {
		return err
	}

	pos.EndedAt = l.pos

	l.curt.Next = NewToken(pos, String, str)
	l.curt = l.curt.Next
	return nil
}

type NumericHandler struct{}

func (n NumericHandler) Handle(l *Lexer, opts *LexOptions) error {
	_ = opts
	pos := NewPosition(l.pos, 0)
	var num []rune
	dotFound := false

	// 負の値?
	if r, ok := l.consume('-'); ok {
		num = append(num, r)
	}

	// .で始まってたら不正な構文
	if _, ok := l.consume('.'); ok {
		return fmt.Errorf("numeric started with dot: at=%v", l.pos)
	}

numLoop:
	for !l.isEof() {
		if r, ok := l.consume('.'); ok {
			// dot
			if dotFound == false {
				// まだdotが含まれていない
				dotFound = true
				num = append(num, r)
			} else {
				// すでにdotが含まれている
				return fmt.Errorf("numeric include multiple dots: at=%v", l.pos)
			}
		} else if r, ok := l.consume('0', '1', '2', '3', '4', '5', '6', '7', '8', '9'); ok {
			// 0~9
			num = append(num, r)
		} else {
			break numLoop
		}
	}

	// num end with dot?
	if num[len(num)-1] == '.' {
		return fmt.Errorf("numeric ended with dot: at=%v", l.pos)
	}

	pos.EndedAt = l.pos

	// dotが含まれている場合はfloatとして扱う
	var kind TokenKind
	if dotFound {
		kind = Float
	} else {
		kind = Integer
	}

	l.curt.Next = NewToken(pos, kind, num)
	l.curt = l.curt.Next
	return nil
}

type SymbolHandler struct{}

func (s SymbolHandler) Handle(l *Lexer, opts *LexOptions) error {
	_ = opts
	pos := NewPosition(l.pos, 0)

	var sym []rune
	var kind TokenKind

	switch {
	case l.charIs('{'):
		r, _ := l.consume('{')
		sym = append(sym, r)
		kind = Lcb
	case l.charIs('}'):
		r, _ := l.consume('}')
		sym = append(sym, r)
		kind = Rcb
	case l.charIs('='):
		r, _ := l.consume('=')
		sym = append(sym, r)
		kind = Assign
	default:
		return fmt.Errorf("unexpected symbol: at=%v", l.pos)
	}

	pos.EndedAt = l.pos
	l.curt.Next = NewToken(pos, kind, sym)
	l.curt = l.curt.Next
	return nil
}
