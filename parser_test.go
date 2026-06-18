package elispvm

import "testing"

func TestParseBasicExpressions(t *testing.T) {
	exprs, err := Parse(`
		; comment
		(concat "a" "b")
		'("read" "grep")
	`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(exprs) != 2 {
		t.Fatalf("len(exprs) = %d, want 2", len(exprs))
	}

	first, ok := exprs[0].(List)
	if !ok {
		t.Fatalf("first expr type = %T, want List", exprs[0])
	}
	if got := Stringify(first); got != `(concat "a" "b")` {
		t.Fatalf("first = %s", got)
	}

	quoted, ok := exprs[1].(List)
	if !ok {
		t.Fatalf("quoted expr type = %T, want List", exprs[1])
	}
	if got := Stringify(quoted); got != `(quote ("read" "grep"))` {
		t.Fatalf("quoted = %s", got)
	}
}

func TestParseStringEscapes(t *testing.T) {
	exprs, err := Parse(`"a\n\"b\""`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	got, ok := exprs[0].(String)
	if !ok {
		t.Fatalf("expr type = %T, want String", exprs[0])
	}
	if string(got) != "a\n\"b\"" {
		t.Fatalf("string = %q", string(got))
	}
}

func TestParseErrors(t *testing.T) {
	tests := []string{
		`(concat "a"`,
		`)`,
		`'`,
		`"unterminated`,
	}
	for _, src := range tests {
		t.Run(src, func(t *testing.T) {
			if _, err := Parse(src); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
