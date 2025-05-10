package parser

import (
	"errors"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexer_Lex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Token
		err   error
	}{
		{
			"ident",
			"on_game_start",
			&Token{
				Pos:  NewPosition(0, 13),
				Kind: Ident,
				Raw:  []rune("on_game_start"),
				Next: &Token{
					Pos:  NewPosition(13, 13),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"string",
			"\"hello,world\"",
			&Token{
				Pos:  NewPosition(0, 13),
				Kind: String,
				Raw:  []rune("hello,world"),
				Next: &Token{
					Pos:  NewPosition(13, 13),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"string/no closing quote",
			"\"unclosed",
			nil,
			errors.New("string without closing dq: at=9"),
		},
		{
			"numeric/integer/positive",
			"123",
			&Token{
				Pos:  NewPosition(0, 3),
				Kind: Integer,
				Raw:  []rune("123"),
				Next: &Token{
					Pos:  NewPosition(3, 3),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"numeric/integer/negative",
			"-123",
			&Token{
				Pos:  NewPosition(0, 4),
				Kind: Integer,
				Raw:  []rune("-123"),
				Next: &Token{
					Pos:  NewPosition(4, 4),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"numeric/float/positive",
			"123.4",
			&Token{
				Pos:  NewPosition(0, 5),
				Kind: Float,
				Raw:  []rune("123.4"),
				Next: &Token{
					Pos:  NewPosition(5, 5),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"numeric/float/negative",
			"-123.4",
			&Token{
				Pos:  NewPosition(0, 6),
				Kind: Float,
				Raw:  []rune("-123.4"),
				Next: &Token{
					Pos:  NewPosition(6, 6),
					Kind: Eof,
					Raw:  nil,
					Next: nil,
				},
			},
			nil,
		},
		{
			"numeric/start with dot",
			"-.123",
			nil,
			errors.New("numeric started with dot: at=2"),
		},
		{
			"numeric/end with dot",
			"123.",
			nil,
			errors.New("numeric ended with dot: at=4"),
		},
		{
			"numeric/multiple dots",
			"1.2.3",
			nil,
			errors.New("numeric include multiple dots: at=4"),
		},
		{
			"symbol/{}=",
			"{}=",
			&Token{
				Pos:  NewPosition(0, 1),
				Kind: Lcb,
				Raw:  []rune("{"),
				Next: &Token{
					Pos:  NewPosition(1, 2),
					Kind: Rcb,
					Raw:  []rune("}"),
					Next: &Token{
						Pos:  NewPosition(2, 3),
						Kind: Assign,
						Raw:  []rune("="),
						Next: &Token{
							Pos:  NewPosition(3, 3),
							Kind: Eof,
							Raw:  nil,
							Next: nil,
						},
					},
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := &Lexer{}
			opts := &LexOptions{Clean: false}
			token, err := lexer.Lex(opts, tt.input)
			assert.Equal(t, tt.err, err)
			if diff := cmp.Diff(token, tt.want); diff != "" {
				t.Errorf("value missmatch\n%s", diff)
			}
		})
	}
}
