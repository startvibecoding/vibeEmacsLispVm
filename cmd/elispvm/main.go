package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	elispvm "github.com/startvibecoding/vibeEmacsLispVm"
	"github.com/startvibecoding/vibeEmacsLispVm/internal/repl"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	evalSrc := flag.String("eval", "", "evaluate source and exit")
	filePath := flag.String("file", "", "evaluate a source file and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version)
		return 0
	}
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "unexpected arguments: %v\n", flag.Args())
		flag.Usage()
		return 2
	}
	if *evalSrc != "" && *filePath != "" {
		fmt.Fprintln(os.Stderr, "-eval and -file cannot be used together")
		return 2
	}

	evaluator := elispvm.New()
	ctx := context.Background()
	if *evalSrc != "" {
		return evalAndPrint(ctx, evaluator, *evalSrc)
	}
	if *filePath != "" {
		src, err := os.ReadFile(*filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read file: %v\n", err)
			return 1
		}
		return evalAndPrint(ctx, evaluator, string(src))
	}

	shell := repl.New(evaluator, os.Stdin, os.Stdout, os.Stderr, repl.DefaultConfig())
	if err := shell.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "repl: %v\n", err)
		return 1
	}
	return 0
}

func evalAndPrint(ctx context.Context, evaluator *elispvm.Evaluator, src string) int {
	value, err := evaluator.EvalString(ctx, src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintln(os.Stdout, elispvm.Stringify(value))
	return 0
}
