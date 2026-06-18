package repl

import (
	"bytes"
	"context"
	"strings"
	"testing"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
)

func TestRunEvaluatesExpressions(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	r := New(elispvm.New(), strings.NewReader("(concat \"a\" \"b\")\n:q\n"), &out, &errOut, Config{})

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := out.String(); !strings.Contains(got, `"ab"`) {
		t.Fatalf("output = %q, want evaluated result", got)
	}
	if errOut.Len() != 0 {
		t.Fatalf("err output = %q", errOut.String())
	}
}

func TestInputComplete(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		complete bool
		wantErr  bool
	}{
		{name: "atom", src: "name", complete: true},
		{name: "list", src: "(concat \"a\" \"b\")", complete: true},
		{name: "multi line list", src: "(concat\n \"a\")", complete: true},
		{name: "open list", src: "(concat", complete: false},
		{name: "string paren ignored", src: "(concat \"(\")", complete: true},
		{name: "comment paren ignored", src: "(concat \"a\") ; )", complete: true},
		{name: "extra close", src: ")", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete, err := inputComplete(tt.src)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if complete != tt.complete {
				t.Fatalf("complete = %v, want %v", complete, tt.complete)
			}
		})
	}
}

func TestCompleteLine(t *testing.T) {
	words := []string{"concat", "cond", "format", "length", "let"}

	line, cursor, matches, changed := completeLine([]rune("(con"), 4, words)
	if changed {
		t.Fatal("completeLine() changed = true, want false")
	}
	if got := string(line); got != "(con" {
		t.Fatalf("line = %q, want common prefix unchanged", got)
	}
	if len(matches) != 2 {
		t.Fatalf("matches = %v, want concat and cond", matches)
	}

	line, cursor, _, changed = completeLine([]rune("(form"), len("(form"), words)
	if !changed {
		t.Fatal("completeLine() changed = false, want true")
	}
	if got := string(line); got != "(format" {
		t.Fatalf("line = %q, want completed format", got)
	}
	if cursor != len("(format") {
		t.Fatalf("cursor = %d, want %d", cursor, len("(format"))
	}
}
