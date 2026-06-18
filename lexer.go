package elispvm

import (
	"strconv"
	"unicode"
)

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenLParen
	tokenRParen
	tokenQuote
	tokenString
	tokenAtom
)

type token struct {
	typ tokenType
	lit string
	pos Position
}

type lexer struct {
	src []rune
	i   int
	pos Position
}

func newLexer(src string) *lexer {
	return &lexer{
		src: []rune(src),
		pos: Position{Line: 1, Column: 1},
	}
}

func (l *lexer) next() (token, error) {
	l.skipSpaceAndComments()
	start := l.pos
	if l.i >= len(l.src) {
		return token{typ: tokenEOF, pos: start}, nil
	}

	ch := l.peek()
	switch ch {
	case '(':
		l.advance()
		return token{typ: tokenLParen, lit: "(", pos: start}, nil
	case ')':
		l.advance()
		return token{typ: tokenRParen, lit: ")", pos: start}, nil
	case '\'':
		l.advance()
		return token{typ: tokenQuote, lit: "'", pos: start}, nil
	case '"':
		return l.readString()
	default:
		return l.readAtom()
	}
}

func (l *lexer) skipSpaceAndComments() {
	for l.i < len(l.src) {
		ch := l.peek()
		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}
		if ch == ';' {
			for l.i < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		return
	}
}

func (l *lexer) readString() (token, error) {
	start := l.pos
	l.advance() // opening quote
	out := make([]rune, 0)
	for l.i < len(l.src) {
		ch := l.advance()
		if ch == '"' {
			return token{typ: tokenString, lit: string(out), pos: start}, nil
		}
		if ch == '\\' {
			if l.i >= len(l.src) {
				return token{}, newError(start, "unterminated string")
			}
			esc := l.advance()
			switch esc {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case '\\', '"':
				out = append(out, esc)
			default:
				return token{}, newError(l.pos, "unsupported escape sequence \\%c", esc)
			}
			continue
		}
		out = append(out, ch)
	}
	return token{}, newError(start, "unterminated string")
}

func (l *lexer) readAtom() (token, error) {
	start := l.pos
	out := make([]rune, 0)
	for l.i < len(l.src) {
		ch := l.peek()
		if unicode.IsSpace(ch) || ch == '(' || ch == ')' || ch == '\'' || ch == ';' {
			break
		}
		out = append(out, l.advance())
	}
	if len(out) == 0 {
		return token{}, newError(start, "unexpected character %q", l.peek())
	}
	return token{typ: tokenAtom, lit: string(out), pos: start}, nil
}

func (l *lexer) peek() rune {
	return l.src[l.i]
}

func (l *lexer) advance() rune {
	ch := l.src[l.i]
	l.i++
	if ch == '\n' {
		l.pos.Line++
		l.pos.Column = 1
	} else {
		l.pos.Column++
	}
	return ch
}

func parseAtom(lit string) Value {
	if lit == "nil" {
		return Nil
	}
	if n, err := strconv.ParseFloat(lit, 64); err == nil {
		return Number(n)
	}
	return Symbol(lit)
}
