package repl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var errInterrupted = errors.New("interrupted")

type lineEditor struct {
	in     *os.File
	out    io.Writer
	reader *bufio.Reader
	words  func() []string
}

func newLineEditor(in *os.File, out io.Writer, words func() []string) *lineEditor {
	return &lineEditor{
		in:     in,
		out:    out,
		reader: bufio.NewReader(in),
		words:  words,
	}
}

func (e *lineEditor) readLine(prompt string) (string, bool, error) {
	fmt.Fprint(e.out, prompt)

	var line []rune
	cursor := 0
	for {
		ch, _, err := e.reader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return "", true, nil
			}
			return "", false, err
		}

		switch ch {
		case '\r', '\n':
			fmt.Fprint(e.out, "\r\n")
			return string(line), false, nil
		case '\t':
			next, nextCursor, matches, changed := completeLine(line, cursor, e.words())
			if changed {
				line = next
				cursor = nextCursor
				redrawLine(e.out, prompt, line, cursor)
				continue
			}
			if len(matches) > 1 {
				fmt.Fprintf(e.out, "\r\n%s\r\n", strings.Join(matches, "  "))
				redrawLine(e.out, prompt, line, cursor)
				continue
			}
			fmt.Fprint(e.out, "\a")
		case 3:
			fmt.Fprint(e.out, "^C\r\n")
			return "", false, errInterrupted
		case 4:
			if len(line) == 0 {
				fmt.Fprint(e.out, "\r\n")
				return "", true, nil
			}
		case 8, 127:
			if cursor > 0 {
				line = append(line[:cursor-1], line[cursor:]...)
				cursor--
				redrawLine(e.out, prompt, line, cursor)
			}
		case 27:
			e.readEscape(&line, &cursor, prompt)
		default:
			if ch >= 32 {
				line = append(line[:cursor], append([]rune{ch}, line[cursor:]...)...)
				cursor++
				redrawLine(e.out, prompt, line, cursor)
			}
		}
	}
}

func (e *lineEditor) readEscape(line *[]rune, cursor *int, prompt string) {
	next, _, err := e.reader.ReadRune()
	if err != nil || next != '[' {
		return
	}
	key, _, err := e.reader.ReadRune()
	if err != nil {
		return
	}
	switch key {
	case 'C':
		if *cursor < len(*line) {
			*cursor = *cursor + 1
			redrawLine(e.out, prompt, *line, *cursor)
		}
	case 'D':
		if *cursor > 0 {
			*cursor = *cursor - 1
			redrawLine(e.out, prompt, *line, *cursor)
		}
	}
}

func redrawLine(out io.Writer, prompt string, line []rune, cursor int) {
	fmt.Fprintf(out, "\r\033[K%s%s", prompt, string(line))
	if trailing := len(line) - cursor; trailing > 0 {
		fmt.Fprintf(out, "\033[%dD", trailing)
	}
}
