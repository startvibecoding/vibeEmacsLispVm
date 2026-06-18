package vm

import "testing"

func TestLexerBasicTokens(t *testing.T) {
	cases := []struct {
		src  string
		typs []tokenType
	}{
		{"(", []tokenType{tokenLParen, tokenEOF}},
		{")", []tokenType{tokenRParen, tokenEOF}},
		{"'", []tokenType{tokenQuote, tokenEOF}},
		{"`", []tokenType{tokenBackquote, tokenEOF}},
		{",", []tokenType{tokenComma, tokenEOF}},
		{",@", []tokenType{tokenCommaSplice, tokenEOF}},
		{"foo", []tokenType{tokenAtom, tokenEOF}},
		{`"str"`, []tokenType{tokenString, tokenEOF}},
		{"", []tokenType{tokenEOF}},
	}
	for _, tt := range cases {
		t.Run(tt.src, func(t *testing.T) {
			l := newLexer(tt.src)
			for i, wantTyp := range tt.typs {
				tok, err := l.next()
				if err != nil {
					t.Fatalf("next[%d] error = %v", i, err)
				}
				if tok.typ != wantTyp {
					t.Fatalf("next[%d] typ = %d, want %d", i, tok.typ, wantTyp)
				}
			}
		})
	}
}

func TestLexerStringContent(t *testing.T) {
	l := newLexer(`"hello world"`)
	tok, err := l.next()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if tok.typ != tokenString {
		t.Fatalf("typ = %d, want tokenString", tok.typ)
	}
	if tok.lit != "hello world" {
		t.Fatalf("lit = %q, want %q", tok.lit, "hello world")
	}
}

func TestLexerStringEscapes(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{`"a\nb"`, "a\nb"},
		{`"a\tb"`, "a\tb"},
		{`"a\rb"`, "a\rb"},
		{`"a\\b"`, "a\\b"},
		{`"a\"b"`, "a\"b"},
	}
	for _, tt := range cases {
		t.Run(tt.src, func(t *testing.T) {
			l := newLexer(tt.src)
			tok, err := l.next()
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if tok.lit != tt.want {
				t.Fatalf("lit = %q, want %q", tok.lit, tt.want)
			}
		})
	}
}

func TestLexerStringInvalidEscape(t *testing.T) {
	l := newLexer(`"a\zb"`)
	_, err := l.next()
	if err == nil {
		t.Fatal("expected error for invalid escape")
	}
}

func TestLexerUnterminatedString(t *testing.T) {
	l := newLexer(`"abc`)
	_, err := l.next()
	if err == nil {
		t.Fatal("expected error for unterminated string")
	}
}

func TestLexerAtomContent(t *testing.T) {
	l := newLexer("foo-bar")
	tok, err := l.next()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if tok.typ != tokenAtom {
		t.Fatalf("typ = %d, want tokenAtom", tok.typ)
	}
	if tok.lit != "foo-bar" {
		t.Fatalf("lit = %q, want %q", tok.lit, "foo-bar")
	}
}

func TestLexerWhitespaceAndComments(t *testing.T) {
	l := newLexer("  ; comment\n  foo  ")
	tok, err := l.next()
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if tok.typ != tokenAtom || tok.lit != "foo" {
		t.Fatalf("tok = {%d, %q}, want {tokenAtom, foo}", tok.typ, tok.lit)
	}
}

func TestLexerMultipleTokens(t *testing.T) {
	l := newLexer("(foo bar)")
	expected := []struct {
		typ tokenType
		lit string
	}{
		{tokenLParen, "("},
		{tokenAtom, "foo"},
		{tokenAtom, "bar"},
		{tokenRParen, ")"},
		{tokenEOF, ""},
	}
	for i, want := range expected {
		tok, err := l.next()
		if err != nil {
			t.Fatalf("next[%d] error = %v", i, err)
		}
		if tok.typ != want.typ {
			t.Fatalf("next[%d] typ = %d, want %d", i, tok.typ, want.typ)
		}
		if tok.lit != want.lit {
			t.Fatalf("next[%d] lit = %q, want %q", i, tok.lit, want.lit)
		}
	}
}

func TestLexerPosition(t *testing.T) {
	l := newLexer("foo\nbar")
	tok1, _ := l.next()
	if tok1.pos.Line != 1 || tok1.pos.Column != 1 {
		t.Fatalf("tok1 pos = %v, want {1,1}", tok1.pos)
	}
	tok2, _ := l.next()
	if tok2.pos.Line != 2 || tok2.pos.Column != 1 {
		t.Fatalf("tok2 pos = %v, want {2,1}", tok2.pos)
	}
}

func TestLexerCommaVsCommaSplice(t *testing.T) {
	l := newLexer(",x ,@y")
	tok1, _ := l.next()
	if tok1.typ != tokenComma {
		t.Fatalf("tok1 typ = %d, want tokenComma", tok1.typ)
	}
	tok2, _ := l.next()
	if tok2.typ != tokenCommaSplice {
		t.Fatalf("tok2 typ = %d, want tokenCommaSplice", tok2.typ)
	}
}

func TestParseAtom(t *testing.T) {
	cases := []struct {
		lit  string
		want string // type name
	}{
		{"nil", "NilType"},
		{"42", "Number"},
		{"3.14", "Number"},
		{"-5", "Number"},
		{"foo", "Symbol"},
		{":key", "Symbol"},
	}
	for _, tt := range cases {
		t.Run(tt.lit, func(t *testing.T) {
			v := parseAtom(tt.lit)
			switch tt.want {
			case "NilType":
				if _, ok := v.(NilType); !ok {
					t.Fatalf("type = %T, want NilType", v)
				}
			case "Number":
				if _, ok := v.(Number); !ok {
					t.Fatalf("type = %T, want Number", v)
				}
			case "Symbol":
				if _, ok := v.(Symbol); !ok {
					t.Fatalf("type = %T, want Symbol", v)
				}
			}
		})
	}
}

func TestLexerAtomStopsAtDelimiters(t *testing.T) {
	cases := []struct {
		src  string
		want string
	}{
		{"foo(bar", "foo"},
		{"foo)bar", "foo"},
		{"foo'bar", "foo"},
		{"foo`bar", "foo"},
		{"foo,bar", "foo"},
		{"foo;bar", "foo"},
		{"foo bar", "foo"},
	}
	for _, tt := range cases {
		t.Run(tt.src, func(t *testing.T) {
			l := newLexer(tt.src)
			tok, err := l.next()
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if tok.lit != tt.want {
				t.Fatalf("lit = %q, want %q", tok.lit, tt.want)
			}
		})
	}
}
