package repl

import (
	"sort"
	"strings"
	"unicode"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
)

func completionWords(evaluator *elispvm.Evaluator) []string {
	seen := map[string]struct{}{
		"nil": {},
		"t":   {},
	}
	for _, name := range evaluator.FuncNames() {
		seen[name] = struct{}{}
	}
	for _, name := range evaluator.SpecialNames() {
		seen[name] = struct{}{}
	}
	for _, name := range evaluator.GlobalNames() {
		seen[name] = struct{}{}
	}

	words := make([]string, 0, len(seen))
	for word := range seen {
		words = append(words, word)
	}
	sort.Strings(words)
	return words
}

func completeLine(line []rune, cursor int, words []string) ([]rune, int, []string, bool) {
	start := tokenStart(line, cursor)
	prefix := string(line[start:cursor])
	if prefix == "" {
		return line, cursor, nil, false
	}

	matches := matchingWords(prefix, words)
	if len(matches) == 0 {
		return line, cursor, nil, false
	}

	replacement := matches[0]
	if len(matches) > 1 {
		replacement = commonPrefix(matches)
	}
	if replacement == prefix {
		return line, cursor, matches, false
	}

	next := make([]rune, 0, len(line)-cursor+start+len([]rune(replacement)))
	next = append(next, line[:start]...)
	next = append(next, []rune(replacement)...)
	next = append(next, line[cursor:]...)
	return next, start + len([]rune(replacement)), matches, true
}

func matchingWords(prefix string, words []string) []string {
	matches := make([]string, 0)
	for _, word := range words {
		if strings.HasPrefix(word, prefix) {
			matches = append(matches, word)
		}
	}
	return matches
}

func commonPrefix(words []string) string {
	if len(words) == 0 {
		return ""
	}
	prefix := []rune(words[0])
	for _, word := range words[1:] {
		next := []rune(word)
		limit := len(prefix)
		if len(next) < limit {
			limit = len(next)
		}
		i := 0
		for i < limit && prefix[i] == next[i] {
			i++
		}
		prefix = prefix[:i]
		if len(prefix) == 0 {
			return ""
		}
	}
	return string(prefix)
}

func tokenStart(line []rune, cursor int) int {
	if cursor > len(line) {
		cursor = len(line)
	}
	for i := cursor - 1; i >= 0; i-- {
		if isTokenBoundary(line[i]) {
			return i + 1
		}
	}
	return 0
}

func isTokenBoundary(ch rune) bool {
	return unicode.IsSpace(ch) || ch == '(' || ch == ')' || ch == '\'' || ch == '"' || ch == ';'
}
