package vm

import "testing"

func TestParseEmpty(t *testing.T) {
	exprs, err := Parse("")
	if err != nil {
		t.Fatalf("Parse(\"\") error = %v", err)
	}
	if len(exprs) != 0 {
		t.Fatalf("len(exprs) = %d, want 0", len(exprs))
	}
}

func TestParseWhitespaceOnly(t *testing.T) {
	exprs, err := Parse("   \n\t  ")
	if err != nil {
		t.Fatalf("Parse whitespace error = %v", err)
	}
	if len(exprs) != 0 {
		t.Fatalf("len(exprs) = %d, want 0", len(exprs))
	}
}

func TestParseCommentsOnly(t *testing.T) {
	exprs, err := Parse("; comment\n; another")
	if err != nil {
		t.Fatalf("Parse comments error = %v", err)
	}
	if len(exprs) != 0 {
		t.Fatalf("len(exprs) = %d, want 0", len(exprs))
	}
}

func TestParseAtomSymbol(t *testing.T) {
	exprs, err := Parse("foo")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("len = %d, want 1", len(exprs))
	}
	if got, ok := exprs[0].(Symbol); !ok || string(got) != "foo" {
		t.Fatalf("expr = %v (%T), want Symbol foo", exprs[0], exprs[0])
	}
}

func TestParseAtomKeyword(t *testing.T) {
	exprs, err := Parse(":keyword")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if got, ok := exprs[0].(Symbol); !ok || string(got) != ":keyword" {
		t.Fatalf("expr = %v (%T), want Symbol :keyword", exprs[0], exprs[0])
	}
}

func TestParseAtomNil(t *testing.T) {
	exprs, err := Parse("nil")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if _, ok := exprs[0].(NilType); !ok {
		t.Fatalf("expr type = %T, want NilType", exprs[0])
	}
}

func TestParseAtomT(t *testing.T) {
	exprs, err := Parse("t")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if got, ok := exprs[0].(Symbol); !ok || string(got) != "t" {
		t.Fatalf("expr = %v, want Symbol t", exprs[0])
	}
}

func TestParseNumber(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{"42", 42},
		{"0", 0},
		{"-5", -5},
		{"3.14", 3.14},
		{"-2.5", -2.5},
		{"1e3", 1000},
	}
	for _, tt := range cases {
		t.Run(tt.src, func(t *testing.T) {
			exprs, err := Parse(tt.src)
			if err != nil {
				t.Fatalf("Parse error = %v", err)
			}
			got, ok := exprs[0].(Number)
			if !ok {
				t.Fatalf("type = %T, want Number", exprs[0])
			}
			if float64(got) != tt.want {
				t.Fatalf("value = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	exprs, err := Parse(`"hello"`)
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if got, ok := exprs[0].(String); !ok || string(got) != "hello" {
		t.Fatalf("expr = %v, want String hello", exprs[0])
	}
}

func TestParseStringEscapes(t *testing.T) {
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
			exprs, err := Parse(tt.src)
			if err != nil {
				t.Fatalf("Parse error = %v", err)
			}
			if got := string(exprs[0].(String)); got != tt.want {
				t.Fatalf("value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStringInvalidEscape(t *testing.T) {
	_, err := Parse(`"a\zb"`)
	if err == nil {
		t.Fatal("expected error for invalid escape")
	}
}

func TestParseEmptyString(t *testing.T) {
	exprs, err := Parse(`""`)
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if got := string(exprs[0].(String)); got != "" {
		t.Fatalf("value = %q, want empty", got)
	}
}

func TestParseEmptyList(t *testing.T) {
	exprs, err := Parse("()")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list, ok := exprs[0].(List)
	if !ok {
		t.Fatalf("type = %T, want List", exprs[0])
	}
	if len(list) != 0 {
		t.Fatalf("len = %d, want 0", len(list))
	}
}

func TestParseSimpleList(t *testing.T) {
	exprs, err := Parse("(foo bar)")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list, ok := exprs[0].(List)
	if !ok {
		t.Fatalf("type = %T, want List", exprs[0])
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if string(list[0].(Symbol)) != "foo" {
		t.Fatalf("list[0] = %v, want foo", list[0])
	}
	if string(list[1].(Symbol)) != "bar" {
		t.Fatalf("list[1] = %v, want bar", list[1])
	}
}

func TestParseNestedList(t *testing.T) {
	exprs, err := Parse("(a (b (c)))")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list := exprs[0].(List)
	if len(list) != 2 {
		t.Fatalf("outer len = %d, want 2", len(list))
	}
	inner := list[1].(List)
	if len(inner) != 2 {
		t.Fatalf("inner len = %d, want 2", len(inner))
	}
	inner2 := inner[1].(List)
	if len(inner2) != 1 {
		t.Fatalf("inner2 len = %d, want 1", len(inner2))
	}
}

func TestParseQuote(t *testing.T) {
	exprs, err := Parse("'x")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list, ok := exprs[0].(List)
	if !ok {
		t.Fatalf("type = %T, want List", exprs[0])
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if string(list[0].(Symbol)) != "quote" {
		t.Fatalf("list[0] = %v, want quote", list[0])
	}
}

func TestParseQuoteList(t *testing.T) {
	exprs, err := Parse("'(a b)")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list := exprs[0].(List)
	if string(list[0].(Symbol)) != "quote" {
		t.Fatalf("head = %v, want quote", list[0])
	}
	inner := list[1].(List)
	if len(inner) != 2 {
		t.Fatalf("inner len = %d, want 2", len(inner))
	}
}

func TestParseBackquote(t *testing.T) {
	exprs, err := Parse("`(a ,b ,@c)")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list, ok := exprs[0].(List)
	if !ok || len(list) != 2 {
		t.Fatalf("expr = %v, want (backquote ...)", exprs[0])
	}
	if string(list[0].(Symbol)) != "backquote" {
		t.Fatalf("head = %v, want backquote", list[0])
	}
}

func TestParseMultipleExpressions(t *testing.T) {
	exprs, err := Parse("a b c")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if len(exprs) != 3 {
		t.Fatalf("len = %d, want 3", len(exprs))
	}
}

func TestParseWithComments(t *testing.T) {
	exprs, err := Parse("; comment\na\n; another\nb")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if len(exprs) != 2 {
		t.Fatalf("len = %d, want 2", len(exprs))
	}
}

// Error cases
func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unterminated list", "(a b"},
		{"unexpected )", ")"},
		{"unterminated string", `"abc`},
		{"quote EOF", "'"},
		{"backquote EOF", "`"},
		{"comma EOF", ","},
		{"comma-splice EOF", ",@"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.src)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseMultipleListsWithWhitespace(t *testing.T) {
	exprs, err := Parse("  (a)  \n  (b c)  ")
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	if len(exprs) != 2 {
		t.Fatalf("len = %d, want 2", len(exprs))
	}
}

func TestParseMixedTypes(t *testing.T) {
	exprs, err := Parse(`(1 "two" three nil)`)
	if err != nil {
		t.Fatalf("Parse error = %v", err)
	}
	list := exprs[0].(List)
	if len(list) != 4 {
		t.Fatalf("len = %d, want 4", len(list))
	}
	if _, ok := list[0].(Number); !ok {
		t.Fatalf("list[0] type = %T, want Number", list[0])
	}
	if _, ok := list[1].(String); !ok {
		t.Fatalf("list[1] type = %T, want String", list[1])
	}
	if _, ok := list[2].(Symbol); !ok {
		t.Fatalf("list[2] type = %T, want Symbol", list[2])
	}
	if _, ok := list[3].(NilType); !ok {
		t.Fatalf("list[3] type = %T, want NilType", list[3])
	}
}
