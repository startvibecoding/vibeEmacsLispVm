//go:build !linux

package repl

type terminalState struct{}

func isTerminal(fd int) bool {
	return false
}

func makeRaw(fd int) (*terminalState, error) {
	return nil, nil
}

func restoreTerminal(fd int, state *terminalState) error {
	return nil
}
