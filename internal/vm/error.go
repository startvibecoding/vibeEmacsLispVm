package vm

import "fmt"

// Position identifies a location in source text.
type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Error is a parse or evaluation error with optional source position.
type Error struct {
	Pos Position
	Msg string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Pos.Line > 0 {
		return fmt.Sprintf("%s: %s", e.Pos, e.Msg)
	}
	return e.Msg
}

func newError(pos Position, format string, args ...any) *Error {
	return &Error{Pos: pos, Msg: fmt.Sprintf(format, args...)}
}
