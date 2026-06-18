package main

import (
	"context"
	"fmt"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
)

func main() {
	e := elispvm.New()
	e.RegisterFunc("join", func(ctx *elispvm.EvalContext, args []elispvm.Value) (elispvm.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("join expects 2 arguments")
		}
		left, ok := args[0].(elispvm.String)
		if !ok {
			return nil, fmt.Errorf("join first argument must be a string")
		}
		right, ok := args[1].(elispvm.String)
		if !ok {
			return nil, fmt.Errorf("join second argument must be a string")
		}
		return elispvm.String(string(left) + "/" + string(right)), nil
	})

	value, err := e.EvalString(context.Background(), `(join "phase" "agent")`)
	if err != nil {
		panic(err)
	}
	fmt.Println(elispvm.Stringify(value))
}
