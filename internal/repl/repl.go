package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
)

// Config controls the interactive shell presentation.
type Config struct {
	Prompt             string
	ContinuationPrompt string
	Banner             string
}

// DefaultConfig returns the standard REPL configuration.
func DefaultConfig() Config {
	return Config{
		Prompt:             "elispvm> ",
		ContinuationPrompt: "      -> ",
		Banner:             `vibeEmacsLispVm REPL. Type :help for help, :quit to exit.`,
	}
}

// REPL reads Lisp expressions, evaluates them, and prints their values.
type REPL struct {
	evaluator *elispvm.Evaluator
	in        io.Reader
	out       io.Writer
	errOut    io.Writer
	config    Config
}

// New creates a REPL over the provided streams.
func New(evaluator *elispvm.Evaluator, in io.Reader, out io.Writer, errOut io.Writer, config Config) *REPL {
	if evaluator == nil {
		evaluator = elispvm.New()
	}
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	if config.Prompt == "" {
		config.Prompt = "elispvm> "
	}
	if config.ContinuationPrompt == "" {
		config.ContinuationPrompt = "      -> "
	}
	return &REPL{
		evaluator: evaluator,
		in:        in,
		out:       out,
		errOut:    errOut,
		config:    config,
	}
}

// Run starts the read-eval-print loop.
func (r *REPL) Run(ctx context.Context) error {
	if r.in == nil {
		return fmt.Errorf("repl input is nil")
	}
	if r.config.Banner != "" {
		fmt.Fprintln(r.out, r.config.Banner)
	}

	inFile, inIsFile := r.in.(*os.File)
	outFile, outIsFile := r.out.(*os.File)
	if inIsFile && outIsFile && isTerminal(int(inFile.Fd())) && isTerminal(int(outFile.Fd())) {
		return r.runInteractive(ctx, inFile)
	}

	return r.runScanner(ctx)
}

func (r *REPL) runInteractive(ctx context.Context, in *os.File) error {
	state, err := makeRaw(int(in.Fd()))
	if err != nil {
		return err
	}
	defer restoreTerminal(int(in.Fd()), state)

	editor := newLineEditor(in, r.out, func() []string {
		return completionWords(r.evaluator)
	})

	var src strings.Builder
	prompt := r.config.Prompt
	for {
		line, eof, err := editor.readLine(prompt)
		if err != nil {
			if errors.Is(err, errInterrupted) {
				src.Reset()
				prompt = r.config.Prompt
				continue
			}
			return err
		}
		if eof {
			break
		}

		quit := r.handleLine(ctx, line, &src, &prompt)
		if quit {
			return nil
		}
	}

	if src.Len() > 0 {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (r *REPL) runScanner(ctx context.Context) error {
	scanner := bufio.NewScanner(r.in)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var src strings.Builder
	prompt := r.config.Prompt
	for {
		fmt.Fprint(r.out, prompt)
		if !scanner.Scan() {
			break
		}

		if r.handleLine(ctx, scanner.Text(), &src, &prompt) {
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if src.Len() > 0 {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (r *REPL) handleLine(ctx context.Context, line string, src *strings.Builder, prompt *string) bool {
	if src.Len() == 0 {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ";") {
			return false
		}
		switch trimmed {
		case "":
			return false
		case ":q", ":quit", "(quit)", "(exit)":
			return true
		case ":h", ":help":
			printHelp(r.out)
			return false
		}
	}

	if src.Len() > 0 {
		src.WriteByte('\n')
	}
	src.WriteString(line)

	complete, err := inputComplete(src.String())
	if err != nil {
		fmt.Fprintf(r.errOut, "parse error: %v\n", err)
		src.Reset()
		*prompt = r.config.Prompt
		return false
	}
	if !complete {
		*prompt = r.config.ContinuationPrompt
		return false
	}

	value, err := r.evaluator.EvalString(ctx, src.String())
	if err != nil {
		fmt.Fprintf(r.errOut, "error: %v\n", err)
	} else {
		fmt.Fprintln(r.out, elispvm.Stringify(value))
	}

	src.Reset()
	*prompt = r.config.Prompt
	return false
}

func printHelp(out io.Writer) {
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  :help, :h   show this help")
	fmt.Fprintln(out, "  :quit, :q   exit")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Enter one Lisp expression at a time. Multi-line lists are supported.")
}

func inputComplete(src string) (bool, error) {
	depth := 0
	inString := false
	escaped := false
	inComment := false
	seenCode := false

	for _, ch := range src {
		if inComment {
			if ch == '\n' {
				inComment = false
			}
			continue
		}
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		switch ch {
		case ';':
			inComment = true
		case '"':
			inString = true
			seenCode = true
		case '(':
			depth++
			seenCode = true
		case ')':
			depth--
			if depth < 0 {
				return false, fmt.Errorf("unexpected )")
			}
			seenCode = true
		default:
			if !unicode.IsSpace(ch) {
				seenCode = true
			}
		}
	}

	if inString || escaped {
		return false, nil
	}
	return seenCode && depth == 0, nil
}
